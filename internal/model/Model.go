package model

import "time"

type Category struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"not null"`
	Type      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
