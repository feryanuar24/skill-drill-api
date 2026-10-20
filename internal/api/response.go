package api

import (
	"skill-drill-api/internal/models"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data,omitempty"`
	Pagination *models.Pagination `json:"pagination,omitempty"`
	Errors     interface{}        `json:"errors,omitempty"`
}

func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendPaginatedSuccess(c *gin.Context, statusCode int, message string, data interface{}, pagination *models.Pagination) {
	c.JSON(statusCode, Response{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

func SendError(c *gin.Context, statusCode int, message string, errors interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}
