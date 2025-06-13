package service

import (
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"time"
)

type TransactionService interface {
	Create(userID string, input *model.Transaction) error
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

func (s *txService) Create(userID string, input *model.Transaction) error {
	input.UserID = userID
	input.CreatedAt = time.Now()
	return s.repo.Create(input)
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
