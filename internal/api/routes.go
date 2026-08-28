package api

import (
	"net/http"
	"skill-drill-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// Public routes
	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/api/auth/register", middleware.RecaptchaMiddleware(), h.Register)
	r.POST("/api/auth/login", middleware.RecaptchaMiddleware(), h.Login)
	r.POST("/api/auth/logout", h.Logout)
	r.GET("/api/auth/verify", h.VerifyEmail)
	r.POST("/api/auth/forgot-password", middleware.RecaptchaMiddleware(), h.ForgotPassword)
	r.POST("/api/auth/reset-password", middleware.RecaptchaMiddleware(), h.ResetPassword)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	// Admin routes
	protected.POST("/admin/generate-questions", h.GenerateQuestions)

	// Study routes
	protected.POST("/study/submit", h.Submit)

	// User routes
	protected.POST("/users", h.CreateUser)
	protected.GET("/users", h.GetUsers)
	protected.GET("/users/:id", h.GetUserByID)
	protected.PUT("/users/:id", h.UpdateUser)
	protected.DELETE("/users/:id", h.DeleteUser)

	// Topic routes
	protected.POST("/topics", h.CreateTopic)
	protected.GET("/topics", h.GetTopics)
	protected.GET("/topics/:id", h.GetTopicByID)
	protected.PUT("/topics/:id", h.UpdateTopic)
	protected.DELETE("/topics/:id", h.DeleteTopic)
}
