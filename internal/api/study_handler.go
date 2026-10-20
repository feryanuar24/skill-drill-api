package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) (uint, error) {
	if val, exists := c.Get("user_id"); exists {
		switch v := val.(type) {
		case float64:
			if v > 0 {
				return uint(v), nil
			}
		case uint:
			if v > 0 {
				return v, nil
			}
		case int:
			if v > 0 {
				return uint(v), nil
			}
		case int64:
			if v > 0 {
				return uint(v), nil
			}
		case string:
			id, err := strconv.ParseUint(v, 10, 32)
			if err == nil && id > 0 {
				return uint(id), nil
			}
		}
	}

	for _, headerName := range []string{"X-User-Id", "User-Id", "user_id"} {
		if val := c.GetHeader(headerName); val != "" {
			id, err := strconv.ParseUint(val, 10, 32)
			if err == nil && id > 0 {
				return uint(id), nil
			}
		}
	}

	return 0, fmt.Errorf("user ID not found")
}

func (h *Handler) Submit(c *gin.Context) {
	var req services.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		SendError(c, http.StatusUnauthorized, "Unauthorized: user ID not found", nil)
		return
	}

	correct, user, err := h.Study.Submit(c.Request.Context(), userID, req.QuestionID, req.Answer)
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
