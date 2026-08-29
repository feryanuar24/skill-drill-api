package api

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

const maxPDFSize = 10 * 1024 * 1024

func (h *Handler) GenerateQuestions(c *gin.Context) {
	var req services.GenerateQuestionsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, FormatValidationError(err))
		return
	}

	var reader io.ReaderAt
	var size int64

	file, fileErr := c.FormFile("file_pdf")
	if fileErr == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_pdf must be a PDF document (.pdf)"})
			return
		}

		if file.Size > maxPDFSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_pdf size exceeds maximum allowed limit of 10MB"})
			return
		}

		uploaded, err := file.Open()
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer uploaded.Close()

		data, err := uploaded.Seek(0, io.SeekEnd)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file size"})
			return
		}

		size = data
		if _, err := uploaded.Seek(0, io.SeekStart); err != nil {
			log.Print(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to reset file pointer"})
			return
		}

		reader = uploaded
	}

	questions, err := h.Questions.Generate(c.Request.Context(), req.TopicID, req.PromptTheme, req.NumQuestions, reader, size)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"questions": questions})
}
