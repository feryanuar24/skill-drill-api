package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GenerateQuestions(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.PostForm("topic_id"), 10, 32)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic_id is required"})
		return
	}

	count, err := strconv.Atoi(c.PostForm("num_questions"))
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "num_questions is required"})
		return
	}

	var reader interface {
		ReadAt([]byte, int64) (int, error)
	}
	var size int64

	file, fileErr := c.FormFile("file_pdf")
	if fileErr == nil {
		uploaded, err := file.Open()
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer uploaded.Close()

		data, err := uploaded.Seek(0, 2)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		size = data
		if _, err := uploaded.Seek(0, 0); err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		reader = uploaded
	}

	var questions interface{}

	if reader != nil {
		questions, err = h.Questions.Generate(c, uint(topicID), c.PostForm("prompt_theme"), count, reader, size)
	} else {
		questions, err = h.Questions.Generate(c, uint(topicID), c.PostForm("prompt_theme"), count, nil, 0)
	}
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"questions": questions})
}
