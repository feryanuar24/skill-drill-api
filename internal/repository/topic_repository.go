package repository

import (
	"context"
	"skill-drill-api/internal/models"
)

func (r *Repository) TopicExists(ctx context.Context, topicID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Topic{}).Where("id = ?", topicID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateTopic(ctx context.Context, topic *models.Topic) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

func (r *Repository) GetTopics(ctx context.Context) ([]models.Topic, error) {
	var topics []models.Topic
	err := r.db.WithContext(ctx).Find(&topics).Error
	return topics, err
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
