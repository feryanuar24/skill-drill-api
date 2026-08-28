package models

import "time"

type User struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	Username               string     `gorm:"size:100;not null" json:"username"`
	Email                  string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password               string     `gorm:"size:255;not null" json:"-"`
	FCMToken               *string    `gorm:"size:500" json:"fcm_token,omitempty"`
	CurrentStreak          int        `gorm:"not null;default:0" json:"current_streak"`
	HighestStreak          int        `gorm:"not null;default:0" json:"highest_streak"`
	LastStudyDate          *time.Time `json:"last_study_date,omitempty"`
	IsVerified             bool       `gorm:"not null;default:false" json:"is_verified"`
	VerificationToken      *string    `gorm:"size:255" json:"-"`
	ResetPasswordToken     *string    `gorm:"size:255" json:"-"`
	ResetPasswordExpiresAt *time.Time `json:"-"`
	CreatedAt              time.Time  `json:"created_at"`
}
