package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"skill-drill-api/internal/api"
	"skill-drill-api/internal/middleware"
	"skill-drill-api/internal/repository"
	"skill-drill-api/internal/services"
	"skill-drill-api/internal/worker"
	"skill-drill-api/pkg/fcm"
	"skill-drill-api/pkg/llm"
	"skill-drill-api/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Print("Error loading .env file, using environment variables instead")
	}

	// Set up logging
	cleanupLogging := initLogging()
	defer cleanupLogging()

	// Set up database
	db := initDB()

	// Set up repository
	repo := repository.New(db)

	// Set up timezone
	location := initTimeZone()

	// Set up LLM client
	llmClient := &llm.HTTPClient{
		URL:    os.Getenv("LLM_API_URL"),
		APIKey: os.Getenv("LLM_API_KEY"),
		Model:  os.Getenv("LLM_MODEL"),
		HTTP:   http.DefaultClient,
	}

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
	r.Use(middleware.CORSMiddleware())
	handler.RegisterRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	addr := host + ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	ln4, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("API listening on IPv4 %s (http://127.0.0.1:%s)", addr, port)

	if host == "" || host == "0.0.0.0" {
		if ln6, err := net.Listen("tcp6", "[::1]:"+port); err == nil {
			log.Printf("API also listening on IPv6 [::1]:%s (http://localhost:%s)", port, port)
			go func() {
				_ = srv.Serve(ln6)
			}()
		}
	}

	if err := srv.Serve(ln4); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
