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
	PromptTheme  string `form:"prompt_theme" binding:"omitempty,max=255"`
}

func (s *QuestionService) Generate(ctx context.Context, topicID uint, theme string, count int, reader io.ReaderAt, size int64) ([]models.Question, error) {
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("num_questions must be between 1 and 100")
	}

	topic, err := s.Repo.GetTopicByID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("topic %d not found", topicID)
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Anda adalah pembuat soal ujian yang ahli dan profesional.\n")
	promptBuilder.WriteString(fmt.Sprintf("Buat %d soal pilihan ganda baru dengan materi utama topik: \"%s\".\n", count, topic.Name))

	if strings.TrimSpace(topic.Description) != "" {
		promptBuilder.WriteString(fmt.Sprintf("Deskripsi / ruang lingkup topik: %s\n", strings.TrimSpace(topic.Description)))
	}

	trimmedTheme := strings.TrimSpace(theme)
	if trimmedTheme != "" {
		promptBuilder.WriteString(fmt.Sprintf("Keterangan / petunjuk tambahan (tingkat kesulitan, pola/tipe soal, fokus materi): %s\n", trimmedTheme))
	} else {
		promptBuilder.WriteString("Tingkat kesulitan soal: menengah hingga sulit dengan analisis mendalam.\n")
	}

	// Fetch existing questions under this topic to maintain pattern consistency
	existingQuestions, _, err := s.Repo.GetQuestions(ctx, repository.QuestionFilter{TopicID: topicID, Limit: 3, SortBy: "id", Order: "desc"})
	if err == nil && len(existingQuestions) > 0 {
		promptBuilder.WriteString("\nPENTING: Pola, struktur pertanyaan, dan gaya bahasa soal yang dihasilkan HARUS KONSISTEN dengan pola soal yang sudah ada pada topik ini. Jadikan contoh soal berikut sebagai referensi pola:\n")
		sampleCount := 3
		if len(existingQuestions) < sampleCount {
			sampleCount = len(existingQuestions)
		}
		for i := 0; i < sampleCount; i++ {
			q := existingQuestions[i]
			promptBuilder.WriteString(fmt.Sprintf("Contoh %d:\n- Pertanyaan: %s\n  A. %s\n  B. %s\n  C. %s\n  D. %s\n  Jawaban Benar: %s\n  Penjelasan: %s\n",
				i+1, q.QuestionText, q.OptionA, q.OptionB, q.OptionC, q.OptionD, q.CorrectAnswer, q.Explanation))
		}
	} else {
		promptBuilder.WriteString("\nPastikan seluruh soal memiliki pola, gaya bahasa, dan kedalaman materi yang konsisten dan relevan untuk topik ini.\n")
	}

	if reader != nil {
		text, err := pdfextractor.ExtractText(reader, size)
		if err != nil {
			return nil, err
		}

		promptBuilder.WriteString(fmt.Sprintf("\nGunakan juga teks dokumen berikut sebagai bahan acuan:\nTEKS:\n%s\n", text))
	}

	promptBuilder.WriteString("\nInstruksi Format dan Output:\n")
	promptBuilder.WriteString("1. Tulis seluruh isi soal, pilihan jawaban (A, B, C, D), dan penjelasan dalam bahasa Indonesia.\n")
	promptBuilder.WriteString("2. Kembalikan HANYA array JSON yang valid tanpa tanda markdown (seperti ```json) atau teks pengantar lainnya.\n")
	promptBuilder.WriteString("3. Format objek JSON: [{\"question_text\": \"...\", \"option_a\": \"...\", \"option_b\": \"...\", \"option_c\": \"...\", \"option_d\": \"...\", \"correct_answer\": \"A/B/C/D\", \"explanation\": \"...\"}]\n")

	prompt := promptBuilder.String()

	response, err := s.LLM.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleaned := strings.TrimSpace(response)
	if start := strings.Index(cleaned, "["); start != -1 {
		if end := strings.LastIndex(cleaned, "]"); end != -1 && end >= start {
			cleaned = cleaned[start : end+1]
		}
	}

	var questions []models.Question

	if err := json.Unmarshal([]byte(cleaned), &questions); err != nil {
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

type GetQuestionsQuery struct {
	Page          int    `form:"page"`
	Limit         int    `form:"limit"`
	TopicID       uint   `form:"topic_id"`
	Search        string `form:"search"`
	CorrectAnswer string `form:"correct_answer"`
	SortBy        string `form:"sort_by"`
	Order         string `form:"order"`
	All           bool   `form:"all"`
}

func (s *QuestionService) GetAll(ctx context.Context, query *GetQuestionsQuery) ([]models.Question, *models.Pagination, error) {
	if query == nil {
		query = &GetQuestionsQuery{}
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

	correct := strings.ToUpper(strings.TrimSpace(query.CorrectAnswer))
	if correct != "" && !strings.Contains("ABCD", correct) {
		correct = ""
	}

	filter := repository.QuestionFilter{
		TopicID:       query.TopicID,
		CorrectAnswer: correct,
		Search:        strings.TrimSpace(query.Search),
		SortBy:        strings.TrimSpace(query.SortBy),
		Order:         strings.TrimSpace(query.Order),
		Page:          page,
		Limit:         limit,
	}

	questions, pagination, err := s.Repo.GetQuestions(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	return questions, &pagination, nil
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
