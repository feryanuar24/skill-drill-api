package api

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

func (h *Handler) CreateQuestion(c *gin.Context) {
	var req services.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	question, err := h.Questions.Create(c.Request.Context(), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusCreated, "Question created successfully", question)
}

func (h *Handler) GetQuestions(c *gin.Context) {
	var topicID uint
	if tIDStr := c.Query("topic_id"); tIDStr != "" {
		if val, err := strconv.ParseUint(tIDStr, 10, 32); err == nil {
			topicID = uint(val)
		}
	}

	search := c.Query("search")
	questions, err := h.Questions.GetAll(c.Request.Context(), topicID, search)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Questions retrieved successfully", questions)
}

func (h *Handler) GetQuestionByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid question ID", nil)
		return
	}
	question, err := h.Questions.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		SendError(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Question retrieved successfully", question)
}

func (h *Handler) UpdateQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid question ID", nil)
		return
	}
	var req services.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	updatedQuestion, err := h.Questions.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Question updated successfully", updatedQuestion)
}

func (h *Handler) DeleteQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid question ID", nil)
		return
	}
	if err := h.Questions.Delete(c.Request.Context(), uint(id)); err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Question deleted successfully", nil)
}
