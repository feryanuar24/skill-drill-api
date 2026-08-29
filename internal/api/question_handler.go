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
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	var reader io.ReaderAt
	var size int64

	file, fileErr := c.FormFile("file_pdf")
	if fileErr == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".pdf" {
			SendError(c, http.StatusBadRequest, "Invalid file format", map[string][]string{"file_pdf": {"file_pdf must be a PDF document (.pdf)"}})
			return
		}

		if file.Size > maxPDFSize {
			SendError(c, http.StatusBadRequest, "File size limit exceeded", map[string][]string{"file_pdf": {"file_pdf size exceeds maximum allowed limit of 10MB"}})
			return
		}

		uploaded, err := file.Open()
		if err != nil {
			log.Print(err)
			SendError(c, http.StatusBadRequest, "Failed to open uploaded file", nil)
			return
		}
		defer uploaded.Close()

		data, err := uploaded.Seek(0, io.SeekEnd)
		if err != nil {
			log.Print(err)
			SendError(c, http.StatusBadRequest, "Failed to read uploaded file size", nil)
			return
		}

		size = data
		if _, err := uploaded.Seek(0, io.SeekStart); err != nil {
			log.Print(err)
			SendError(c, http.StatusBadRequest, "Failed to reset file pointer", nil)
			return
		}

		reader = uploaded
	}

	questions, err := h.Questions.Generate(c.Request.Context(), req.TopicID, req.PromptTheme, req.NumQuestions, reader, size)
	if err != nil {
		log.Print(err)
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendSuccess(c, http.StatusCreated, "Questions generated successfully", gin.H{"questions": questions})
}
