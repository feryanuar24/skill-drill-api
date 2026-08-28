package services

import (
	"context"
	"fmt"
	"strings"

	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *repository.Repository
}

type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	FCMToken *string `json:"fcm_token,omitempty"`
}

type UpdateUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FCMToken *string `json:"fcm_token,omitempty"`
}

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FCMToken: req.FCMToken,
	}

	if err := s.Repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAll(ctx context.Context) ([]models.User, error) {
	return s.Repo.GetUsers(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.Repo.GetUserByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id uint, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.Repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}
	if req.FCMToken != nil {
		user.FCMToken = req.FCMToken
	}

	if err := s.Repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	_, err := s.Repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	return s.Repo.DeleteUser(ctx, id)
}
