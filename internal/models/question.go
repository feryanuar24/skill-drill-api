package models

import "time"

type Question struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TopicID       uint      `gorm:"not null;index" json:"topic_id"`
	Topic         *Topic    `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	QuestionText  string    `gorm:"type:text;not null" json:"question_text"`
	OptionA       string    `gorm:"type:text;not null" json:"option_a"`
	OptionB       string    `gorm:"type:text;not null" json:"option_b"`
	OptionC       string    `gorm:"type:text;not null" json:"option_c"`
	OptionD       string    `gorm:"type:text;not null" json:"option_d"`
	CorrectAnswer string    `gorm:"size:1;not null" json:"correct_answer"`
	Explanation   string    `gorm:"type:text" json:"explanation"`
	CreatedAt     time.Time `json:"created_at"`
}
