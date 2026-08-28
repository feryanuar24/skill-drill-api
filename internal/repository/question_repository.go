package repository

import (
	"context"
	"skill-drill-api/internal/models"
)

func (r *Repository) CreateQuestions(ctx context.Context, questions []models.Question) error {
	return r.db.WithContext(ctx).CreateInBatches(questions, 100).Error
}

func (r *Repository) FindQuestion(ctx context.Context, questionID, topicID uint) (*models.Question, error) {
	var question models.Question
	err := r.db.WithContext(ctx).Where("id = ? AND topic_id = ?", questionID, topicID).First(&question).Error
	return &question, err
}
