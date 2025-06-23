package repository

import (
	"statistic_service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(wallet *model.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *WalletRepository) AddMember(member *model.WalletMember) error {
	return r.db.Create(member).Error
}

func (r *WalletRepository) GetTransactions(walletID uuid.UUID) ([]model.Transaction, error) {
	var txs []model.Transaction
	err := r.db.
		Preload("Category").
		Where("wallet_id = ?", walletID).
		Find(&txs).Error
	return txs, err
}

func (r *WalletRepository) GetMembers(walletID uuid.UUID) ([]model.WalletMember, error) {
	var members []model.WalletMember
	err := r.db.Where("wallet_id = ?", walletID).Find(&members).Error
	return members, err
}

func (r *WalletRepository) AddTransaction(tx *model.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *WalletRepository) GetCategoryByID(id uuid.UUID) (*model.Category, error) {
	var category model.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *WalletRepository) CreateTransaction(tx *model.Transaction) error {
	return r.db.Create(tx).Error
}
