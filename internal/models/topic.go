package models

import "time"

type Topic struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	Questions   []Question `json:"-"`
}
