package repository

import (
	"context"
	"fmt"
	"strings"

	"skill-drill-api/internal/models"

	"gorm.io/gorm"
)

func (r *Repository) CreateQuestion(ctx context.Context, question *models.Question) error {
	return r.db.WithContext(ctx).Create(question).Error
}

func (r *Repository) CreateQuestions(ctx context.Context, questions []models.Question) error {
	return r.db.WithContext(ctx).CreateInBatches(questions, 100).Error
}

func (r *Repository) FindQuestion(ctx context.Context, questionID uint) (*models.Question, error) {
	var question models.Question
	err := r.db.WithContext(ctx).Where("id = ?", questionID).First(&question).Error
	return &question, err
}

type QuestionFilter struct {
	TopicID       uint
	CorrectAnswer string
	Search        string
	SortBy        string
	Order         string
	Page          int
	Limit         int
}

func (r *Repository) GetQuestions(ctx context.Context, filter QuestionFilter) ([]models.Question, models.Pagination, error) {
	query := r.db.WithContext(ctx).Model(&models.Question{}).Preload("Topic")

	if filter.TopicID > 0 {
		query = query.Where("topic_id = ?", filter.TopicID)
	}
	if filter.CorrectAnswer != "" {
		query = query.Where("correct_answer = ?", filter.CorrectAnswer)
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(question_text) LIKE ? OR LOWER(explanation) LIKE ? OR LOWER(option_a) LIKE ? OR LOWER(option_b) LIKE ? OR LOWER(option_c) LIKE ? OR LOWER(option_d) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	var totalItems int64
	if err := query.Session(&gorm.Session{}).Count(&totalItems).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	sortBy := "created_at"
	allowedSorts := map[string]string{
		"id":             "id",
		"topic_id":       "topic_id",
		"question_text":  "question_text",
		"correct_answer": "correct_answer",
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

	var questions []models.Question
	if err := query.Find(&questions).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	pagination := models.CalculatePagination(totalItems, page, limit)
	return questions, pagination, nil
}

func (r *Repository) GetQuestionByID(ctx context.Context, id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.WithContext(ctx).Preload("Topic").First(&question, id).Error
	return &question, err
}

func (r *Repository) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return r.db.WithContext(ctx).Save(question).Error
}

func (r *Repository) DeleteQuestion(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Question{}, id).Error
}
