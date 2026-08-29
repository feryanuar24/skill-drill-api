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
