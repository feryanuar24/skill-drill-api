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
	users, err := h.Users.GetAll(c.Request.Context())
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendSuccess(c, http.StatusOK, "Users retrieved successfully", users)
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
