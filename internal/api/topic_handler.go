package api

import (
	"net/http"
	"strconv"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateTopic(c *gin.Context) {
	var req services.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	topic, err := h.Topics.Create(c.Request.Context(), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusCreated, "Topic created successfully", topic)
}

func (h *Handler) GetTopics(c *gin.Context) {
	var query services.GetTopicsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		SendError(c, http.StatusBadRequest, "Invalid query parameters", FormatValidationError(err))
		return
	}

	if query.Search == "" {
		query.Search = c.Query("q")
	}
	if query.SortBy == "" {
		query.SortBy = c.Query("sort")
	}
	if query.Order == "" {
		query.Order = c.Query("order_by")
	}
	if query.Limit == 0 {
		if val, err := strconv.Atoi(c.Query("per_page")); err == nil && val > 0 {
			query.Limit = val
		} else if val, err := strconv.Atoi(c.Query("page_size")); err == nil && val > 0 {
			query.Limit = val
		}
	}
	if !query.All && c.Query("all") == "true" {
		query.All = true
	}

	topics, pagination, err := h.Topics.GetAll(c.Request.Context(), &query)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendPaginatedSuccess(c, http.StatusOK, "Topics retrieved successfully", topics, pagination)
}

func (h *Handler) GetTopicByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid topic ID", nil)
		return
	}
	topic, err := h.Topics.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		SendError(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Topic retrieved successfully", topic)
}

func (h *Handler) UpdateTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid topic ID", nil)
		return
	}
	var req services.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	updatedTopic, err := h.Topics.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Topic updated successfully", updatedTopic)
}

func (h *Handler) DeleteTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid topic ID", nil)
		return
	}
	if err := h.Topics.Delete(c.Request.Context(), uint(id)); err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Topic deleted successfully", nil)
}
