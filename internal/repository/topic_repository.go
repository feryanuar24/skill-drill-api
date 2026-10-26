package repository

import (
	"context"
	"fmt"
	"strings"

	"skill-drill-api/internal/models"

	"gorm.io/gorm"
)

func (r *Repository) TopicExists(ctx context.Context, topicID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Topic{}).Where("id = ?", topicID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateTopic(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

type TopicFilter struct {
	Search string
	SortBy string
	Order  string
	Page   int
	Limit  int
}

func (r *Repository) GetTopics(ctx context.Context, filter TopicFilter) ([]models.Topic, models.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&models.Topic{})

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	var totalItems int64
	if err := query.Session(&gorm.Session{}).Count(&totalItems).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	sortBy := "created_at"
	allowedSorts := map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"created_at":  "created_at",
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

	var topics []models.Topic
	if err := query.Find(&topics).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	pagination := models.CalculatePagination(totalItems, page, limit)
	return topics, pagination, nil
}

func (r *Repository) GetTopicByID(ctx context.Context, id uint) (*models.Topic, error) {
	var topic models.Topic
	err := r.db.WithContext(ctx).First(&topic, id).Error
	return &topic, err
}

func (r *Repository) UpdateTopic(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Save(topic).Error
}

func (r *Repository) DeleteTopic(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Topic{}, id).Error
}
