package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type Category struct {
	ID     uuid.UUID `db:"id"      json:"id"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	Name   string    `db:"name"    json:"name"`
	Type   string    `db:"type"    json:"type"` // "expense" немесе "income"
}

type Transaction struct {
	ID         uuid.UUID       `db:"id"          json:"id"`
	UserID     uuid.UUID       `db:"user_id"     json:"user_id"`
	Amount     decimal.Decimal `db:"amount"      json:"amount"`
	Type       string          `db:"type"        json:"type"`
	CategoryID *uuid.UUID      `db:"category_id" json:"category_id"` // null болуы мумкин
	Comment    *string         `db:"comment"     json:"comment"`     // null болуы мумкин
	CreatedAt  time.Time       `db:"created_at"  json:"created_at"`
}
