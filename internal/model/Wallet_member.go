package model

import "time"

// WalletMember представляет связь между пользователем и кошельком,
// а также определяет роль пользователя в этом кошельке.
type WalletMember struct {
	WalletID  string    `gorm:"primaryKey;type:uuid;not null" json:"wallet_id"` // ID кошелька
	UserID    string    `gorm:"primaryKey;type:uuid;not null" json:"user_id"`   // ID пользователя
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`          // Роль: "admin", "member"
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи для удобства загрузки данных
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"-"` // Скрываем, чтобы избежать циклических зависимостей при JSON-маршалинге
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Определения ролей
const (
	WalletRoleAdmin  = "admin"
	WalletRoleMember = "member"
)
