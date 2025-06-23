package model

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
	Members   []WalletMember `gorm:"foreignKey:WalletID"`
}

type WalletMember struct {
	WalletID uuid.UUID `gorm:"primaryKey"`
	UserID   uuid.UUID `gorm:"primaryKey"`
	Role     string
}
