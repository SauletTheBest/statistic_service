// internal/model/Transaction.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `json:"user_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // "income" or "expense"
	Description string    `json:"description"`
	// Старое поле: Category    string    `json:"category"`
	CategoryID uuid.UUID `json:"category_id"`                           // <-- НОВОЕ ПОЛЕ
	Category   Category  `json:"category" gorm:"foreignKey:CategoryID"` // Связь с моделью Category
	Date       time.Time `json:"date"`
	CreatedAt  time.Time `json:"created_at"`
}
