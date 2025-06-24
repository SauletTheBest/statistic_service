package model

import "time"

type RefreshToken struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"not null"`
	Token     string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
