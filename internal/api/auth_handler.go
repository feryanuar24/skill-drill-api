package api

import (
	"log"
	"net/http"
	"strings"

	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	res, err := h.Auth.Register(c.Request.Context(), req)
	if err != nil {
		log.Print(err)
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendSuccess(c, http.StatusCreated, "User registered successfully", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	res, err := h.Auth.Login(c.Request.Context(), req)
	if err != nil {
		log.Print(err)
		SendError(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	c.SetCookie("token", res.Token, 3600*72, "/", "", false, true)

	SendSuccess(c, http.StatusOK, "Login successful", res)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	SendSuccess(c, http.StatusOK, "Logged out successfully", nil)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		SendError(c, http.StatusBadRequest, "Verification token is required", nil)
		return
	}

	if err := h.Auth.VerifyEmail(c.Request.Context(), token); err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendSuccess(c, http.StatusOK, "Email verified successfully", nil)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req services.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	if err := h.Auth.ForgotPassword(c.Request.Context(), req); err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to process request", nil)
		return
	}

	SendSuccess(c, http.StatusOK, "If the email exists, a reset link has been sent", nil)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Validation failed", FormatValidationError(err))
		return
	}

	if err := h.Auth.ResetPassword(c.Request.Context(), req); err != nil {
		SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendSuccess(c, http.StatusOK, "Password reset successfully", nil)
}
