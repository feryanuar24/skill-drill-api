package api

import (
	"skill-drill-api/internal/services"
)

type Handler struct {
	Questions *services.QuestionService
	Study     *services.StudyService
	Users     *services.UserService
	Topics    *services.TopicService
	Auth      *services.AuthService
}
