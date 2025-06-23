package service

import (
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type TransactionService interface {
	Create(userID uuid.UUID, walletID uuid.UUID, amount float64, categoryID *uuid.UUID, comment string) (*model.Transaction, error)
	List(userID string, from, to *time.Time, txType string) ([]model.Transaction, error)
	Update(id string, userID string, input *model.Transaction) error
	Delete(id, userID string) error
	Summary(userID string, from, to *time.Time) (float64, float64, error)
	ByCategory(userID string, from, to *time.Time) (map[string]float64, error)
}

type txService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(r repository.TransactionRepository) TransactionService {
	return &txService{r}
}

func (s *txService) Create(userID uuid.UUID, walletID uuid.UUID, amount float64, categoryID *uuid.UUID, comment string) (*model.Transaction, error) {
	tx := &model.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID.String(),
		WalletID:   &walletID,
		Amount:     amount,
		CategoryID: categoryID,
		Comment:    comment,
		Type:       "expense", // если нужно — передавай как аргумент
		CreatedAt:  time.Now(),
	}
	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *txService) List(userID string, from, to *time.Time, txType string) ([]model.Transaction, error) {
	return s.repo.GetByUser(userID, from, to, txType)
}
func (s *txService) Update(id, userID string, input *model.Transaction) error {
	existing, err := s.repo.GetByID(id)
	if err != nil || existing.UserID != userID {
		return err
	}
	existing.Amount = input.Amount
	existing.Type = input.Type
	existing.Category = input.Category
	existing.Comment = input.Comment
	return s.repo.Update(existing)
}
func (s *txService) Delete(id, userID string) error {
	tx, err := s.repo.GetByID(id)
	if err != nil || tx.UserID != userID {
		return err
	}
	return s.repo.Delete(id)
}
func (s *txService) Summary(userID string, from, to *time.Time) (float64, float64, error) {
	return s.repo.Summary(userID, from, to)
}
func (s *txService) ByCategory(userID string, from, to *time.Time) (map[string]float64, error) {
	return s.repo.ByCategory(userID, from, to)
}
