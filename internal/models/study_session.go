package models

import "time"

type StudySession struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	TopicID     uint      `gorm:"not null;index" json:"topic_id"`
	Score       int       `gorm:"not null;default:0" json:"score"`
	IsCompleted bool      `gorm:"not null;default:false" json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
}
