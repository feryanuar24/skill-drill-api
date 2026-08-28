package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"skill-drill-api/internal/api"
	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
	"skill-drill-api/internal/services"
	"skill-drill-api/internal/worker"
	"skill-drill-api/pkg/fcm"
	"skill-drill-api/pkg/llm"

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

	// Set up API handler
	handler := &api.Handler{
		Questions: &services.QuestionService{Repo: repo, LLM: llmClient},
		Study:     &services.StudyService{Repo: repo, Location: location},
		Users:     &services.UserService{Repo: repo},
		Topics:    &services.TopicService{Repo: repo},
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
	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/api/admin/generate-questions", handler.GenerateQuestions)
	r.POST("/api/study/submit", handler.Submit)

	// User routes
	r.POST("/api/users", handler.CreateUser)
	r.GET("/api/users", handler.GetUsers)
	r.GET("/api/users/:id", handler.GetUserByID)
	r.PUT("/api/users/:id", handler.UpdateUser)
	r.DELETE("/api/users/:id", handler.DeleteUser)

	// Topic routes
	r.POST("/api/topics", handler.CreateTopic)
	r.GET("/api/topics", handler.GetTopics)
	r.GET("/api/topics/:id", handler.GetTopicByID)
	r.PUT("/api/topics/:id", handler.UpdateTopic)
	r.DELETE("/api/topics/:id", handler.DeleteTopic)

	// Start server
	log.Printf("API listening on :%s", os.Getenv("PORT"))
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
