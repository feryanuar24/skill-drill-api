package repository

import (
	"context"
	"skill-drill-api/internal/models"
)

func (r *Repository) CreateSession(ctx context.Context, session *models.StudySession) error {
	return r.db.WithContext(ctx).Create(session).Error
}
