package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"skill-drill-api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initLogging() func() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logWriter)
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter

	return func() {
		logFile.Close()
	}
}

func initDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("TIMEZONE"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Question{}, &models.StudySession{}); err != nil {
		log.Fatal(err)
	}
	return db
}

func initTimeZone() *time.Location {
	location, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		log.Fatal(err)
	}
	return location
}
