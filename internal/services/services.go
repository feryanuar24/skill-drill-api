package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"skill-drill-api/internal/models"
	"skill-drill-api/internal/repository"
	"skill-drill-api/pkg/llm"
	pdfextractor "skill-drill-api/pkg/pdf"
)

type QuestionService struct {
	Repo *repository.Repository
	LLM  llm.Client
}

func (s *QuestionService) Generate(ctx context.Context, topicID uint, theme string, count int, reader io.ReaderAt, size int64) ([]models.Question, error) {
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("num_questions must be between 1 and 100")
	}

	if strings.TrimSpace(theme) == "" {
		return nil, fmt.Errorf("prompt_theme is required")
	}

	exists, err := s.Repo.TopicExists(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("topic %d not found", topicID)
	}

	prompt := fmt.Sprintf("Anda adalah pembuat soal yang ahli. Buat %d soal pilihan ganda acak dengan tingkat kesulitan menengah hingga sulit tentang %s. Tulis seluruh isi soal, pilihan jawaban, dan penjelasan dalam bahasa Indonesia. Kembalikan HANYA array JSON yang valid dengan objek berisi question_text, option_a, option_b, option_c, option_d, correct_answer (A, B, atau C, atau D), dan explanation.", count, theme)
	if reader != nil {
		text, err := pdfextractor.ExtractText(reader, size)
		if err != nil {
			return nil, err
		}

		prompt = fmt.Sprintf("Berdasarkan teks berikut, buat %d soal pilihan ganda tentang %s. Tulis seluruh isi soal, pilihan jawaban, dan penjelasan dalam bahasa Indonesia. Kembalikan HANYA array JSON yang valid dengan objek berisi question_text, option_a, option_b, option_c, option_d, correct_answer (A, B, atau C, atau D), dan explanation.\n\nTEKS:\n%s", count, theme, text)
	}

	response, err := s.LLM.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var questions []models.Question

	if err := json.Unmarshal([]byte(response), &questions); err != nil {
		return nil, fmt.Errorf("parse llm JSON array: %w", err)
	}

	if len(questions) != count {
		return nil, fmt.Errorf("llm returned %d questions, expected %d", len(questions), count)
	}

	for i := range questions {
		questions[i].ID = 0
		questions[i].TopicID = topicID
		questions[i].CreatedAt = time.Time{}

		if !strings.Contains("ABCD", strings.ToUpper(questions[i].CorrectAnswer)) {
			return nil, fmt.Errorf("invalid correct_answer at index %d", i)
		}
	}

	if err := s.Repo.CreateQuestions(ctx, questions); err != nil {
		return nil, err
	}

	return questions, nil
}

type StudyService struct {
	Repo     *repository.Repository
	Location *time.Location
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

type UserService struct {
	Repo *repository.Repository
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	return s.Repo.CreateUser(ctx, user)
}

func (s *UserService) GetAll(ctx context.Context) ([]models.User, error) {
	return s.Repo.GetUsers(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.Repo.GetUserByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id uint, req *models.User) (*models.User, error) {
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
