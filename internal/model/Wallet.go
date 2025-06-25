package model

import "time"

// Wallet представляет собой кошелек, которым могут делиться несколько пользователей.
type Wallet struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OwnerID   string    `gorm:"type:uuid;not null;index" json:"owner_id"` // ID пользователя, создавшего кошелек
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связи
	Members      []WalletMember `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE;" json:"members,omitempty"`
	Transactions []Transaction  `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE;" json:"transactions,omitempty"`
}
