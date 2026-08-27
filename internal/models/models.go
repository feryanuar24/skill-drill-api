package models

import "time"

type User struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Username      string     `gorm:"size:100;not null" json:"username"`
	Email         string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	FCMToken      *string    `gorm:"size:500" json:"fcm_token,omitempty"`
	CurrentStreak int        `gorm:"not null;default:0" json:"current_streak"`
	HighestStreak int        `gorm:"not null;default:0" json:"highest_streak"`
	LastStudyDate *time.Time `json:"last_study_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Topic struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	Questions   []Question `json:"-"`
}

type Question struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TopicID       uint      `gorm:"not null;index" json:"topic_id"`
	QuestionText  string    `gorm:"type:text;not null" json:"question_text"`
	OptionA       string    `gorm:"type:text;not null" json:"option_a"`
	OptionB       string    `gorm:"type:text;not null" json:"option_b"`
	OptionC       string    `gorm:"type:text;not null" json:"option_c"`
	OptionD       string    `gorm:"type:text;not null" json:"option_d"`
	CorrectAnswer string    `gorm:"size:1;not null" json:"correct_answer"`
	Explanation   string    `gorm:"type:text" json:"explanation"`
	CreatedAt     time.Time `json:"created_at"`
}

type StudySession struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	TopicID     uint      `gorm:"not null;index" json:"topic_id"`
	Score       int       `gorm:"not null;default:0" json:"score"`
	IsCompleted bool      `gorm:"not null;default:false" json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
}
