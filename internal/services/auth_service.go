package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
	"skill-drill-api/pkg/mailer"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo   *repository.Repository
	Mailer *mailer.Mailer
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type AuthResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"`
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if _, err := s.Repo.FindUserByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	tokenStr := generateRandomToken()

	user := models.User{
		Username:          req.Username,
		Email:             req.Email,
		Password:          string(hashedPassword),
		IsVerified:        false,
		VerificationToken: &tokenStr,
	}

	if err := s.Repo.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	go func() {
		subject := "Verify your Email for Skill Drill API"
		body := fmt.Sprintf("Hi %s,\n\nPlease verify your email by using this token: %s", user.Username, tokenStr)
		_ = s.Mailer.SendEmail(user.Email, subject, body)
	}()

	return &AuthResponse{User: user}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.Repo.FindUserByVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid verification token")
	}

	user.IsVerified = true
	user.VerificationToken = nil

	if err := s.Repo.SaveUser(ctx, user); err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.Repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsVerified {
		return nil, errors.New("email not verified")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{User: *user, Token: token}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := s.Repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	tokenStr := generateRandomToken()
	expiry := time.Now().Add(1 * time.Hour)

	user.ResetPasswordToken = &tokenStr
	user.ResetPasswordExpiresAt = &expiry

	if err := s.Repo.SaveUser(ctx, user); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	go func() {
		subject := "Reset your Password for Skill Drill API"
		body := fmt.Sprintf("Hi %s,\n\nYou requested a password reset. Use this token to reset your password: %s", user.Username, tokenStr)
		_ = s.Mailer.SendEmail(user.Email, subject, body)
	}()

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	user, err := s.Repo.FindUserByResetPasswordToken(ctx, req.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if user.ResetPasswordExpiresAt == nil || time.Now().After(*user.ResetPasswordExpiresAt) {
		return errors.New("invalid or expired reset token")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordToken = nil
	user.ResetPasswordExpiresAt = nil

	if err := s.Repo.SaveUser(ctx, user); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

func (s *AuthService) generateToken(userID uint) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
