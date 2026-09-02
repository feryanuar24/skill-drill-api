package repository

import (
	"context"
	"skill-drill-api/internal/models"
)

func (r *Repository) CreateQuestion(ctx context.Context, question *models.Question) error {
	return r.db.WithContext(ctx).Create(question).Error
}

func (r *Repository) CreateQuestions(ctx context.Context, questions []models.Question) error {
	return r.db.WithContext(ctx).CreateInBatches(questions, 100).Error
}

func (r *Repository) FindQuestion(ctx context.Context, questionID, topicID uint) (*models.Question, error) {
	var question models.Question
	err := r.db.WithContext(ctx).Where("id = ? AND topic_id = ?", questionID, topicID).First(&question).Error
	return &question, err
}

func (r *Repository) GetQuestions(ctx context.Context, topicID uint, search string) ([]models.Question, error) {
	var questions []models.Question
	query := r.db.WithContext(ctx).Preload("Topic")

	if topicID > 0 {
		query = query.Where("topic_id = ?", topicID)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("question_text LIKE ? OR explanation LIKE ?", searchPattern, searchPattern)
	}

	err := query.Order("id desc").Find(&questions).Error
	return questions, err
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
