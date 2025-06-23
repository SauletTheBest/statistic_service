package service

import (
	"fmt"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) CreateWallet(ownerID uuid.UUID, name string) (*model.Wallet, error) {
	wallet := &model.Wallet{
		Name:    name,
		OwnerID: ownerID,
	}
	if err := s.repo.Create(wallet); err != nil {
		return nil, err
	}
	member := model.WalletMember{
		WalletID: wallet.ID,
		UserID:   ownerID,
		Role:     "admin",
	}
	if err := s.repo.AddMember(&member); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) InviteMember(walletID, userID uuid.UUID, role string) error {
	member := model.WalletMember{
		WalletID: walletID,
		UserID:   userID,
		Role:     role,
	}
	return s.repo.AddMember(&member)
}

func (s *WalletService) GetTransactions(walletID uuid.UUID) ([]model.Transaction, error) {
	txs, err := s.repo.GetTransactions(walletID)
	return txs, err

}

func (s *WalletService) GetMembers(walletID uuid.UUID) ([]model.WalletMember, error) {
	members, err := s.repo.GetMembers(walletID)
	return members, err
}

func (s *WalletService) CreateTransaction(walletID, userID uuid.UUID, amount float64, categoryID *uuid.UUID, comment string) (*model.Transaction, error) {
	var category *model.Category
	if categoryID != nil {
		var err error
		category, err = s.repo.GetCategoryByID(*categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to find category: %w", err)
		}
	}

	tx := &model.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID.String(),
		Amount:     amount,
		CategoryID: categoryID,
		WalletID:   &walletID,
		Comment:    comment,
	}

	if category != nil {
		tx.Type = category.Type // 'income' or 'expense'
	}

	// Не забудь сохранить в БД!
	if err := s.repo.CreateTransaction(tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}
