package repository

import (
	"statistic_service/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WalletRepository определяет интерфейс для операций с кошельками.
type WalletRepository interface {
	Create(wallet *model.Wallet) error
	GetByID(walletID uuid.UUID) (*model.Wallet, error)
	ListByUserID(userID uuid.UUID) ([]model.Wallet, error)
	Update(wallet *model.Wallet) error
	Delete(walletID, ownerID uuid.UUID) error
	CountByOwnerID(ownerID uuid.UUID) (int64, error)

	AddMember(member *model.WalletMember) error
	GetMember(walletID, userID uuid.UUID) (*model.WalletMember, error)
	UpdateMemberRole(walletID, userID uuid.UUID, role string) error
	RemoveMember(walletID, userID uuid.UUID) error
	GetMembers(walletID uuid.UUID) ([]model.WalletMember, error)
	GetTransactionsByWalletID(walletID uuid.UUID, from, to *time.Time, txType string) ([]model.Transaction, error)
}

type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository создает новый экземпляр WalletRepository.
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *model.Wallet) error {
	// Создаем кошелек и участника-владельца в одной транзакции
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(wallet).Error; err != nil {
			return err
		}
		// Владелец автоматически становится админом
		ownerMember := &model.WalletMember{
			WalletID: wallet.ID,
			UserID:   wallet.OwnerID,
			Role:     model.WalletRoleAdmin,
		}
		if err := tx.Create(ownerMember).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *walletRepository) GetByID(walletID uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.Preload("Members.User").First(&wallet, "id = ?", walletID).Error
	return &wallet, err
}

func (r *walletRepository) ListByUserID(userID uuid.UUID) ([]model.Wallet, error) {
	var wallets []model.Wallet
	// Находим все ID кошельков, где пользователь является участником
	err := r.db.Joins("JOIN wallet_members ON wallets.id = wallet_members.wallet_id").
		Where("wallet_members.user_id = ?", userID).
		Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) Update(wallet *model.Wallet) error {
	return r.db.Save(wallet).Error
}

// Удалять может только владелец
func (r *walletRepository) Delete(walletID, ownerID uuid.UUID) error {
	result := r.db.Where("id = ? AND owner_id = ?", walletID, ownerID).Delete(&model.Wallet{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *walletRepository) CountByOwnerID(ownerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Wallet{}).Where("owner_id = ?", ownerID).Count(&count).Error
	return count, err
}

func (r *walletRepository) AddMember(member *model.WalletMember) error {
	return r.db.Create(member).Error
}

func (r *walletRepository) GetMember(walletID, userID uuid.UUID) (*model.WalletMember, error) {
	var member model.WalletMember
	err := r.db.Where("wallet_id = ? AND user_id = ?", walletID, userID).First(&member).Error
	return &member, err
}

func (r *walletRepository) UpdateMemberRole(walletID, userID uuid.UUID, role string) error {
	return r.db.Model(&model.WalletMember{}).
		Where("wallet_id = ? AND user_id = ?", walletID, userID).
		Update("role", role).Error
}

func (r *walletRepository) RemoveMember(walletID, userID uuid.UUID) error {
	result := r.db.Where("wallet_id = ? AND user_id = ?", walletID, userID).Delete(&model.WalletMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *walletRepository) GetMembers(walletID uuid.UUID) ([]model.WalletMember, error) {
	var members []model.WalletMember
	err := r.db.Preload("User").Where("wallet_id = ?", walletID).Find(&members).Error
	return members, err
}

func (r *walletRepository) GetTransactionsByWalletID(walletID uuid.UUID, from, to *time.Time, txType string) ([]model.Transaction, error) {
	var transactions []model.Transaction
	q := r.db.Preload("Category").Where("wallet_id = ?", walletID)
	if txType != "" {
		q = q.Where("type = ?", txType)
	}
	if from != nil {
		q = q.Where("date >= ?", *from)
	}
	if to != nil {
		q = q.Where("date <= ?", *to)
	}
	if err := q.Order("date DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
