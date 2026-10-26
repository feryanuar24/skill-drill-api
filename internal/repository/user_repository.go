package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skill-drill-api/internal/models"

	"gorm.io/gorm"
)

func (r *Repository) FindUser(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	return &user, err
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *Repository) FindUserByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("verification_token = ?", token).First(&user).Error
	return &user, err
}

func (r *Repository) FindUserByResetPasswordToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("reset_password_token = ?", token).First(&user).Error
	return &user, err
}

func (r *Repository) SaveUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) UsersToNotify(ctx context.Context, today time.Time) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("fcm_token IS NOT NULL AND (last_study_date IS NULL OR last_study_date::date <> ?)", today.Format("2006-01-02")).Find(&users).Error
	return users, err
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

type UserFilter struct {
	Search     string
	Role       string
	IsVerified *bool
	SortBy     string
	Order      string
	Page       int
	Limit      int
}

func (r *Repository) GetUsers(ctx context.Context, filter UserFilter) ([]models.User, models.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&models.User{})

	if filter.Role != "" && strings.ToLower(filter.Role) != "all" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.IsVerified != nil {
		query = query.Where("is_verified = ?", *filter.IsVerified)
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}

	var totalItems int64
	if err := query.Session(&gorm.Session{}).Count(&totalItems).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	sortBy := "created_at"
	allowedSorts := map[string]string{
		"id":             "id",
		"username":       "username",
		"email":          "email",
		"role":           "role",
		"is_verified":    "is_verified",
		"current_streak": "current_streak",
		"highest_streak": "highest_streak",
		"created_at":     "created_at",
	}
	if col, ok := allowedSorts[strings.ToLower(filter.SortBy)]; ok {
		sortBy = col
	}

	order := "desc"
	if strings.ToLower(filter.Order) == "asc" {
		order = "asc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit

	if limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	pagination := models.CalculatePagination(totalItems, page, limit)
	return users, pagination, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) DeleteUser(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
