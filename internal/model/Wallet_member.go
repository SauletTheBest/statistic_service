package model

import "time"

// WalletMember представляет связь между пользователем и кошельком.
type WalletMember struct {
	WalletID  string    `gorm:"primaryKey;type:uuid;not null" json:"wallet_id"`
	UserID    string    `gorm:"primaryKey;type:uuid;not null" json:"user_id"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"` // Роли: "admin", "member"
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи для удобства загрузки данных (Preload)
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Определения ролей, которые будут доступны во всем проекте
const (
	WalletRoleAdmin  = "admin"
	WalletRoleMember = "member"
)
