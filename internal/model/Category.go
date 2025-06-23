package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "income" or "expense"
	CreatedAt time.Time `json:"created_at"`
}
