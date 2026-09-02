package api

import (
	"net/http"
	"skill-drill-api/internal/middleware"
	"skill-drill-api/internal/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// Initialize rate limiters
	authLimiter := middleware.AuthRateLimiter()
	genLimiter := middleware.GenQuestionsRateLimiter()

	// Public routes
	r.GET("/api/health", func(c *gin.Context) {
		SendSuccess(c, http.StatusOK, "API is running", nil)
	})
	r.POST("/api/auth/register", authLimiter, middleware.RecaptchaMiddleware(), h.Register)
	r.POST("/api/auth/login", authLimiter, middleware.RecaptchaMiddleware(), h.Login)
	r.POST("/api/auth/logout", h.Logout)
	r.GET("/api/auth/verify", h.VerifyEmail)
	r.POST("/api/auth/forgot-password", authLimiter, middleware.RecaptchaMiddleware(), h.ForgotPassword)
	r.POST("/api/auth/reset-password", authLimiter, middleware.RecaptchaMiddleware(), h.ResetPassword)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	// Admin routes
	admin := protected.Group("")
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
	{
		admin.POST("/admin/generate-questions", genLimiter, h.GenerateQuestions)

		// User routes
		admin.POST("/users", h.CreateUser)
		admin.GET("/users", h.GetUsers)
		admin.GET("/users/:id", h.GetUserByID)
		admin.PUT("/users/:id", h.UpdateUser)
		admin.DELETE("/users/:id", h.DeleteUser)

		// Topic routes
		admin.POST("/topics", h.CreateTopic)
		admin.GET("/topics", h.GetTopics)
		admin.GET("/topics/:id", h.GetTopicByID)
		admin.PUT("/topics/:id", h.UpdateTopic)
		admin.DELETE("/topics/:id", h.DeleteTopic)

		// Question routes
		admin.POST("/questions", h.CreateQuestion)
		admin.GET("/questions", h.GetQuestions)
		admin.GET("/questions/:id", h.GetQuestionByID)
		admin.PUT("/questions/:id", h.UpdateQuestion)
		admin.DELETE("/questions/:id", h.DeleteQuestion)
	}

	// Study routes (accessible by Student & Admin)
	student := protected.Group("")
	student.Use(middleware.RoleMiddleware(models.RoleStudent, models.RoleAdmin))
	{
		student.POST("/study/submit", h.Submit)
	}
}
