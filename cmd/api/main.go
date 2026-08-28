package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"skill-drill-api/internal/api"
	"skill-drill-api/internal/middleware"
	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
	"skill-drill-api/internal/services"
	"skill-drill-api/internal/worker"
	"skill-drill-api/pkg/fcm"
	"skill-drill-api/pkg/llm"
	"skill-drill-api/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file, using environment variables instead")
	}

	// Set up loggging
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	logWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logWriter)
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter

	// Set up database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Question{}, &models.StudySession{}); err != nil {
		log.Fatal(err)
	}
	repo := repository.New(db)

	// Set up timezone
	location, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		log.Fatal(err)
	}

	// Set up LLM client
	llmClient := &llm.HTTPClient{URL: os.Getenv("LLM_API_URL"), APIKey: os.Getenv("LLM_API_KEY"), Model: os.Getenv("LLM_MODEL"), HTTP: http.DefaultClient}

	// Set up Mailer
	appMailer := mailer.NewMailer()

	// Set up API handler
	handler := &api.Handler{
		Questions: &services.QuestionService{Repo: repo, LLM: llmClient},
		Study:     &services.StudyService{Repo: repo, Location: location},
		Users:     &services.UserService{Repo: repo},
		Topics:    &services.TopicService{Repo: repo},
		Auth:      &services.AuthService{Repo: repo, Mailer: appMailer},
	}

	// Set up FCM client
	fcmClient, err := fcm.NewClient(os.Getenv("FIREBASE_CREDENTIALS_FILE"), http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	notifier := (&worker.Notifier{Repo: repo, FCM: fcmClient, Location: location}).Start()
	defer notifier.Stop()

	// Set up Gin router
	r := gin.Default()

	// Public routes
	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/api/auth/register", handler.Register)
	r.POST("/api/auth/login", handler.Login)
	r.GET("/api/auth/verify", handler.VerifyEmail)
	r.POST("/api/auth/forgot-password", handler.ForgotPassword)
	r.POST("/api/auth/reset-password", handler.ResetPassword)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	protected.POST("/admin/generate-questions", handler.GenerateQuestions)
	protected.POST("/study/submit", handler.Submit)

	// User routes
	protected.POST("/users", handler.CreateUser)
	protected.GET("/users", handler.GetUsers)
	protected.GET("/users/:id", handler.GetUserByID)
	protected.PUT("/users/:id", handler.UpdateUser)
	protected.DELETE("/users/:id", handler.DeleteUser)

	// Topic routes
	protected.POST("/topics", handler.CreateTopic)
	protected.GET("/topics", handler.GetTopics)
	protected.GET("/topics/:id", handler.GetTopicByID)
	protected.PUT("/topics/:id", handler.UpdateTopic)
	protected.DELETE("/topics/:id", handler.DeleteTopic)

	// Start server
	log.Printf("API listening on :%s", os.Getenv("PORT"))
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
