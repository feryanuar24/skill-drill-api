package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type submitRequest struct {
	UserID     uint   `json:"user_id"`
	TopicID    uint   `json:"topic_id"`
	QuestionID uint   `json:"question_id"`
	Answer     string `json:"answer"`
}

func (h *Handler) Submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	correct, user, err := h.Study.Submit(c, req.UserID, req.TopicID, req.QuestionID, req.Answer)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"correct": correct, "current_streak": user.CurrentStreak, "highest_streak": user.HighestStreak})
}
