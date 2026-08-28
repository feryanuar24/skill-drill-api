package services

import (
	"context"
	"fmt"
	"strings"

	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
)

type TopicService struct {
	Repo *repository.Repository
}

func (s *TopicService) Create(ctx context.Context, topic *models.Topic) error {
	topic.Name = strings.TrimSpace(topic.Name)
	if topic.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.Repo.CreateTopic(ctx, topic)
}

func (s *TopicService) GetAll(ctx context.Context) ([]models.Topic, error) {
	return s.Repo.GetTopics(ctx)
}

func (s *TopicService) GetByID(ctx context.Context, id uint) (*models.Topic, error) {
	return s.Repo.GetTopicByID(ctx, id)
}

func (s *TopicService) Update(ctx context.Context, id uint, req *models.Topic) (*models.Topic, error) {
	topic, err := s.Repo.GetTopicByID(ctx, id)
	if err != nil {
		return nil, err
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name != "" {
		topic.Name = req.Name
	}
	topic.Description = req.Description

	if err := s.Repo.UpdateTopic(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

func (s *TopicService) Delete(ctx context.Context, id uint) error {
	_, err := s.Repo.GetTopicByID(ctx, id)
	if err != nil {
		return err
	}
	return s.Repo.DeleteTopic(ctx, id)
}
