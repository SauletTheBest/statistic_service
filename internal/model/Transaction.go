package model

import "time"

type Transaction struct {
	ID       string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID   string    `gorm:"type:uuid;not null;index" json:"user_id"` // Пользователь, создавший транзакцию
	WalletID string    `gorm:"type:uuid;index" json:"wallet_id"`        // Убрали `not null`
	Amount   float64   `gorm:"not null" json:"amount"`
	Type     string    `gorm:"not null" json:"type"` // "income" or "expense"
	Comment  string    `json:"Comment"`
	Date     time.Time `gorm:"not null" json:"date"` // Фактическая дата транзакции

	CategoryID string   `gorm:"type:uuid;not null;index" json:"category_id"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"` // Добавляем UpdatedAt
}
