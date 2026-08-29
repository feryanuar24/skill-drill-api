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
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	correct, user, err := h.Study.Submit(c.Request.Context(), req.UserID, req.TopicID, req.QuestionID, req.Answer)
	if err != nil {
		log.Print(err)
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendSuccess(c, http.StatusOK, "Answer submitted successfully", gin.H{
		"correct":        correct,
		"current_streak": user.CurrentStreak,
		"highest_streak": user.HighestStreak,
	})
}
