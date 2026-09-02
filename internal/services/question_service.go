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

type GenerateQuestionsRequest struct {
	TopicID      uint   `form:"topic_id" binding:"required,min=1"`
	NumQuestions int    `form:"num_questions" binding:"required,min=1,max=100"`
	PromptTheme  string `form:"prompt_theme" binding:"required,min=2,max=255"`
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

type CreateQuestionRequest struct {
	TopicID       uint   `json:"topic_id" binding:"required,min=1"`
	QuestionText  string `json:"question_text" binding:"required,min=5"`
	OptionA       string `json:"option_a" binding:"required"`
	OptionB       string `json:"option_b" binding:"required"`
	OptionC       string `json:"option_c" binding:"required"`
	OptionD       string `json:"option_d" binding:"required"`
	CorrectAnswer string `json:"correct_answer" binding:"required,oneof=A B C D a b c d"`
	Explanation   string `json:"explanation"`
}

type UpdateQuestionRequest struct {
	TopicID       uint   `json:"topic_id" binding:"omitempty,min=1"`
	QuestionText  string `json:"question_text" binding:"omitempty,min=5"`
	OptionA       string `json:"option_a" binding:"omitempty"`
	OptionB       string `json:"option_b" binding:"omitempty"`
	OptionC       string `json:"option_c" binding:"omitempty"`
	OptionD       string `json:"option_d" binding:"omitempty"`
	CorrectAnswer string `json:"correct_answer" binding:"omitempty,oneof=A B C D a b c d"`
	Explanation   string `json:"explanation"`
}

func (s *QuestionService) Create(ctx context.Context, req *CreateQuestionRequest) (*models.Question, error) {
	exists, err := s.Repo.TopicExists(ctx, req.TopicID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("topic %d not found", req.TopicID)
	}

	correct := strings.ToUpper(strings.TrimSpace(req.CorrectAnswer))
	if !strings.Contains("ABCD", correct) || len(correct) != 1 {
		return nil, fmt.Errorf("correct_answer must be A, B, C, or D")
	}

	question := &models.Question{
		TopicID:       req.TopicID,
		QuestionText:  strings.TrimSpace(req.QuestionText),
		OptionA:       strings.TrimSpace(req.OptionA),
		OptionB:       strings.TrimSpace(req.OptionB),
		OptionC:       strings.TrimSpace(req.OptionC),
		OptionD:       strings.TrimSpace(req.OptionD),
		CorrectAnswer: correct,
		Explanation:   strings.TrimSpace(req.Explanation),
	}

	if err := s.Repo.CreateQuestion(ctx, question); err != nil {
		return nil, err
	}

	return s.Repo.GetQuestionByID(ctx, question.ID)
}

func (s *QuestionService) GetAll(ctx context.Context, topicID uint, search string) ([]models.Question, error) {
	return s.Repo.GetQuestions(ctx, topicID, strings.TrimSpace(search))
}

func (s *QuestionService) GetByID(ctx context.Context, id uint) (*models.Question, error) {
	return s.Repo.GetQuestionByID(ctx, id)
}

func (s *QuestionService) Update(ctx context.Context, id uint, req *UpdateQuestionRequest) (*models.Question, error) {
	question, err := s.Repo.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.TopicID > 0 && req.TopicID != question.TopicID {
		exists, err := s.Repo.TopicExists(ctx, req.TopicID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("topic %d not found", req.TopicID)
		}
		question.TopicID = req.TopicID
	}

	if strings.TrimSpace(req.QuestionText) != "" {
		question.QuestionText = strings.TrimSpace(req.QuestionText)
	}
	if strings.TrimSpace(req.OptionA) != "" {
		question.OptionA = strings.TrimSpace(req.OptionA)
	}
	if strings.TrimSpace(req.OptionB) != "" {
		question.OptionB = strings.TrimSpace(req.OptionB)
	}
	if strings.TrimSpace(req.OptionC) != "" {
		question.OptionC = strings.TrimSpace(req.OptionC)
	}
	if strings.TrimSpace(req.OptionD) != "" {
		question.OptionD = strings.TrimSpace(req.OptionD)
	}
	if strings.TrimSpace(req.CorrectAnswer) != "" {
		correct := strings.ToUpper(strings.TrimSpace(req.CorrectAnswer))
		if !strings.Contains("ABCD", correct) || len(correct) != 1 {
			return nil, fmt.Errorf("correct_answer must be A, B, C, or D")
		}
		question.CorrectAnswer = correct
	}
	if req.Explanation != "" || req.QuestionText != "" {
		question.Explanation = strings.TrimSpace(req.Explanation)
	}

	if err := s.Repo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}

	return s.Repo.GetQuestionByID(ctx, question.ID)
}

func (s *QuestionService) Delete(ctx context.Context, id uint) error {
	_, err := s.Repo.GetQuestionByID(ctx, id)
	if err != nil {
		return err
	}
	return s.Repo.DeleteQuestion(ctx, id)
}
