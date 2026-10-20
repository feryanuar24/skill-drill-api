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

type CreateTopicRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

type UpdateTopicRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

func (s *TopicService) Create(ctx context.Context, req *CreateTopicRequest) (*models.Topic, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	topic := &models.Topic{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
	}

	if err := s.Repo.CreateTopic(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

type GetTopicsQuery struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
	SortBy string `form:"sort_by"`
	Order  string `form:"order"`
	All    bool   `form:"all"`
}

func (s *TopicService) GetAll(ctx context.Context, query *GetTopicsQuery) ([]models.Topic, *models.Pagination, error) {
	if query == nil {
		query = &GetTopicsQuery{}
	}

	limit := query.Limit
	if query.All {
		limit = 0
	} else if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	page := query.Page
	if page < 1 {
		page = 1
	}

	filter := repository.TopicFilter{
		Search: strings.TrimSpace(query.Search),
		SortBy: strings.TrimSpace(query.SortBy),
		Order:  strings.TrimSpace(query.Order),
		Page:   page,
		Limit:  limit,
	}

	topics, pagination, err := s.Repo.GetTopics(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	return topics, &pagination, nil
}

func (s *TopicService) GetByID(ctx context.Context, id uint) (*models.Topic, error) {
	return s.Repo.GetTopicByID(ctx, id)
}

func (s *TopicService) Update(ctx context.Context, id uint, req *UpdateTopicRequest) (*models.Topic, error) {
	topic, err := s.Repo.GetTopicByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name != "" {
		topic.Name = name
	}
	if req.Description != "" || req.Name != "" {
		topic.Description = strings.TrimSpace(req.Description)
	}

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
