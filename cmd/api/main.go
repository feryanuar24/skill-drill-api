package main

import (
	"fmt"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getenv("DB_HOST", "localhost"),
		getenv("DB_USER", "postgres"),
		getenv("DB_PASSWORD", "password"),
		getenv("DB_NAME", "skill_drill"),
		getenv("DB_PORT", "5432"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Question{}, &models.StudySession{}); err != nil {
		log.Fatal(err)
	}
	repo := repository.New(db)
	location, err := time.LoadLocation(getenv("TIMEZONE", "UTC"))
	if err != nil {
		log.Fatal(err)
	}
	llmClient := &llm.HTTPClient{URL: getenv("LLM_API_URL", "https://api.deepseek.com/chat/completions"), APIKey: os.Getenv("LLM_API_KEY"), Model: getenv("LLM_MODEL", "deepseek-chat"), HTTP: http.DefaultClient}
	handler := &api.Handler{Questions: &services.QuestionService{Repo: repo, LLM: llmClient}, Study: &services.StudyService{Repo: repo, Location: location}}
	fcmClient, err := fcm.NewClient(getenv("FIREBASE_CREDENTIALS_FILE", "firebase.json"), http.DefaultClient)
	if err != nil {
		log.Printf("FCM disabled: %v", err)
	}
	notifier := (&worker.Notifier{Repo: repo, FCM: fcmClient, Location: location}).Start()
	defer notifier.Stop()
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/admin/generate-questions", handler.GenerateQuestions)
	r.POST("/study/submit", handler.Submit)
	log.Printf("API listening on :%s", getenv("PORT", "8080"))
	if err := r.Run(":" + getenv("PORT", "8080")); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
