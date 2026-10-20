package api

import (
	"net/http"
	"strconv"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	user, err := h.Users.Create(c.Request.Context(), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusCreated, "User created successfully", user)
}

func (h *Handler) GetUsers(c *gin.Context) {
	var query services.GetUsersQuery
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

	users, pagination, err := h.Users.GetAll(c.Request.Context(), &query)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendPaginatedSuccess(c, http.StatusOK, "Users retrieved successfully", users, pagination)
}

func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	user, err := h.Users.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		SendError(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "User retrieved successfully", user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}
	updatedUser, err := h.Users.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "User updated successfully", updatedUser)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		SendError(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	if err := h.Users.Delete(c.Request.Context(), uint(id)); err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "User deleted successfully", nil)
}
