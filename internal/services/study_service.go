package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
)

type StudyService struct {
	Repo     *repository.Repository
	Location *time.Location
}

type SubmitAnswerRequest struct {
	UserID     uint   `json:"user_id" binding:"required,min=1"`
	TopicID    uint   `json:"topic_id" binding:"required,min=1"`
	QuestionID uint   `json:"question_id" binding:"required,min=1"`
	Answer     string `json:"answer" binding:"required,oneof=A B C D a b c d"`
}

func (s *StudyService) Submit(ctx context.Context, userID, topicID, questionID uint, answer string) (bool, *models.User, error) {
	answer = strings.ToUpper(strings.TrimSpace(answer))
	if !strings.Contains("ABCD", answer) || len(answer) != 1 {
		return false, nil, fmt.Errorf("answer must be A, B, C, or D")
	}

	question, err := s.Repo.FindQuestion(ctx, questionID, topicID)
	if err != nil {
		return false, nil, err
	}

	user, err := s.Repo.FindUser(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	now := time.Now().In(s.Location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	if user.LastStudyDate == nil {
		user.CurrentStreak = 1
	} else {
		last := user.LastStudyDate.In(s.Location)
		lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, s.Location)
		gap := int(today.Sub(lastDay).Hours() / 24)

		if gap == 1 {
			user.CurrentStreak++
		} else if gap > 1 {
			user.CurrentStreak = 0
		}
	}

	if user.CurrentStreak > user.HighestStreak {
		user.HighestStreak = user.CurrentStreak
	}

	user.LastStudyDate = &today
	if err := s.Repo.SaveUser(ctx, user); err != nil {
		return false, nil, err
	}

	session := &models.StudySession{UserID: userID, TopicID: topicID, Score: 0, IsCompleted: true}
	if answer == strings.ToUpper(question.CorrectAnswer) {
		session.Score = 1
	}

	if err := s.Repo.CreateSession(ctx, session); err != nil {
		return false, nil, err
	}

	return session.Score == 1, user, nil
}
