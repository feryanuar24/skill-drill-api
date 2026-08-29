package api

import (
	"log"
	"net/http"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Submit(c *gin.Context) {
	var req services.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FormatValidationError(err))
		return
	}

	correct, user, err := h.Study.Submit(c.Request.Context(), req.UserID, req.TopicID, req.QuestionID, req.Answer)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"correct":        correct,
		"current_streak": user.CurrentStreak,
		"highest_streak": user.HighestStreak,
	})
}
