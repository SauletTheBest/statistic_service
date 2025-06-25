package service

import (
	"errors"
	"fmt"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

// TransactionService определяет интерфейс для управления транзакциями.
type TransactionService interface {
	Create(userID string, transaction *model.Transaction) error
	List(userID string, startDate, endDate *time.Time, transactionType string) ([]model.Transaction, error)
	Update(userID string, transaction *model.Transaction) error
	Delete(userID, transactionID string) error
	Summary(userID string, transactionType string, startDate, endDate *time.Time) (float64, error)
	ByCategory(userID string, transactionType string, startDate, endDate *time.Time) (map[string]float64, error)
	// Добавляем метод для проверки существования категории
	CheckCategoryExists(categoryID string, userID string) error
}

// TransactionServiceImpl реализует TransactionService.
type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
	categoryRepo    repository.CategoryRepository // Добавим репозиторий категорий
	logger          *logrus.Logger                // <-- НОВОЕ: Добавляем логгер
}

// NewTransactionService создает новый экземпляр TransactionServiceImpl.
// <-- ИСПРАВЛЕНО: Теперь принимает логгер
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	categoryRepo repository.CategoryRepository,
	logger *logrus.Logger, // <-- НОВОЕ
) TransactionService {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		logger:          logger, // <-- НОВОЕ
	}
}

// Create создает новую транзакцию.
func (s *TransactionServiceImpl) Create(userID string, transaction *model.Transaction) error {
	// Проверяем, существует ли категория и принадлежит ли она пользователю
	err := s.CheckCategoryExists(transaction.CategoryID, userID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"userID":     userID,
			"categoryID": transaction.CategoryID,
		}).Warn("Category check failed during transaction creation")
		return err // Вернет "Category not found" или "Category does not belong to user"
	}

	transaction.UserID = userID
	if transaction.Date.IsZero() {
		transaction.Date = time.Now()
	}
	if err := s.transactionRepo.Create(transaction); err != nil {
		s.logger.WithError(err).WithField("userID", userID).Error("Failed to create transaction in repository")
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	s.logger.WithField("transactionID", transaction.ID).Info("Transaction created successfully")
	return nil
}

// List возвращает список транзакций пользователя за указанный период.
func (s *TransactionServiceImpl) List(userID string, startDate, endDate *time.Time, transactionType string) ([]model.Transaction, error) {
	transactions, err := s.transactionRepo.List(userID, startDate, endDate, transactionType)
	if err != nil {
		s.logger.WithError(err).WithField("userID", userID).Error("Failed to list transactions from repository")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return transactions, nil
}

// Update обновляет существующую транзакцию.
func (s *TransactionServiceImpl) Update(userID string, transaction *model.Transaction) error {
	// Проверяем, существует ли категория и принадлежит ли она пользователю (если категория меняется)
	if transaction.CategoryID != "" { // Только если CategoryID был предоставлен для обновления
		err := s.CheckCategoryExists(transaction.CategoryID, userID)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"userID":        userID,
				"transactionID": transaction.ID,
				"newCategoryID": transaction.CategoryID,
			}).Warn("Category check failed during transaction update")
			return err
		}
	}

	// Проверяем, что транзакция принадлежит текущему пользователю
	existingTx, err := s.transactionRepo.GetByID(transaction.ID)
	if err != nil {
		s.logger.WithError(err).WithField("transactionID", transaction.ID).Error("Failed to get transaction by ID for update")
		return fmt.Errorf("transaction not found: %w", err)
	}
	if existingTx.UserID != userID {
		s.logger.WithFields(logrus.Fields{
			"userID":        userID,
			"transactionID": transaction.ID,
			"ownerID":       existingTx.UserID,
		}).Warn("Attempt to update transaction not owned by user")
		return errors.New("transaction does not belong to user")
	}

	transaction.UserID = userID // Убедимся, что UserID не изменится
	if err := s.transactionRepo.Update(transaction); err != nil {
		s.logger.WithError(err).WithField("transactionID", transaction.ID).Error("Failed to update transaction in repository")
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	s.logger.WithField("transactionID", transaction.ID).Info("Transaction updated successfully")
	return nil
}

// Delete удаляет транзакцию по ID.
func (s *TransactionServiceImpl) Delete(userID, transactionID string) error {
	// Проверяем, что транзакция принадлежит текущему пользователю перед удалением
	existingTx, err := s.transactionRepo.GetByID(transactionID)
	if err != nil {
		s.logger.WithError(err).WithField("transactionID", transactionID).Error("Failed to get transaction by ID for deletion")
		return fmt.Errorf("transaction not found: %w", err)
	}
	if existingTx.UserID != userID {
		s.logger.WithFields(logrus.Fields{
			"userID":        userID,
			"transactionID": transactionID,
			"ownerID":       existingTx.UserID,
		}).Warn("Attempt to delete transaction not owned by user")
		return errors.New("transaction does not belong to user")
	}

	if err := s.transactionRepo.Delete(transactionID); err != nil {
		s.logger.WithError(err).WithField("transactionID", transactionID).Error("Failed to delete transaction from repository")
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	s.logger.WithField("transactionID", transactionID).Info("Transaction deleted successfully")
	return nil
}

// Summary рассчитывает общую сумму транзакций за указанный период.
func (s *TransactionServiceImpl) Summary(userID string, transactionType string, startDate, endDate *time.Time) (float64, error) {
	total, err := s.transactionRepo.GetTotalAmount(userID, transactionType, startDate, endDate)
	if err != nil {
		s.logger.WithError(err).WithField("userID", userID).Error("Failed to get transaction summary from repository")
		return 0, fmt.Errorf("failed to get summary: %w", err)
	}
	return total, nil
}

// ByCategory агрегирует суммы транзакций по категориям.
func (s *TransactionServiceImpl) ByCategory(userID string, transactionType string, startDate, endDate *time.Time) (map[string]float64, error) {
	categoryTotals, err := s.transactionRepo.GetTotalAmountByCategory(userID, transactionType, startDate, endDate)
	if err != nil {
		s.logger.WithError(err).WithField("userID", userID).Error("Failed to get transaction totals by category from repository")
		return nil, fmt.Errorf("failed to get transactions by category: %w", err)
	}
	return categoryTotals, nil
}

// CheckCategoryExists проверяет, существует ли категория и принадлежит ли она указанному пользователю.
func (s *TransactionServiceImpl) CheckCategoryExists(categoryID string, userID string) error {
	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"categoryID": categoryID,
			"userID":     userID,
		}).Warn("Category not found or failed to retrieve during check")
		return errors.New("category not found")
	}
	if category.UserID != userID {
		s.logger.WithFields(logrus.Fields{
			"categoryID": categoryID,
			"userID":     userID,
			"ownerID":    category.UserID,
		}).Warn("Category does not belong to user")
		return errors.New("category does not belong to user")
	}
	return nil
}
