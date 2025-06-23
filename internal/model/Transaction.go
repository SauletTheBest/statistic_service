package model

import (
	"time"
)

type Transaction struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	Amount     float64   `gorm:"not null"`
	Type       string    `gorm:"type:text;not null"`
	Category   string    `gorm:"type:text"`
	CategoryID string    `gorm:"type:uuid" json:"category_id,omitempty"`
	Comment    string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

//ewdwdqwdqdqw
