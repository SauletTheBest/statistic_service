package service

import (
	"errors" // Добавляем импорт для errors
	"statistic_service/internal/model"
	"statistic_service/internal/repository"
	"time"
)

// TransactionService определяет интерфейс для операций с транзакциями.
type TransactionService interface {
	Create(userID string, input *model.Transaction) error
	List(userID string, from, to *time.Time, txType string) ([]model.Transaction, error)
	Update(id string, userID string, input *model.Transaction) error
	Delete(id, userID string) error
	Summary(userID string, from, to *time.Time) (float64, float64, error)
	ByCategory(userID string, from, to *time.Time) (map[string]float64, error)
}

// txService реализует TransactionService, используя TransactionRepository и CategoryRepository.
type txService struct {
	repo         repository.TransactionRepository
	categoryRepo repository.CategoryRepository // <-- НОВОЕ: Добавляем CategoryRepository
}

// NewTransactionService создает новый экземпляр TransactionService.
// <-- НОВОЕ: Принимает CategoryRepository
func NewTransactionService(r repository.TransactionRepository, cr repository.CategoryRepository) TransactionService {
	return &txService{repo: r, categoryRepo: cr}
}

// Create создает новую транзакцию.
func (s *txService) Create(userID string, input *model.Transaction) error {
	// Валидация CategoryID
	if input.CategoryID != "" {
		category, err := s.categoryRepo.GetByID(input.CategoryID)
		if err != nil {
			return errors.New("category not found") // Категория не найдена
		}
		if category.UserID != userID {
			return errors.New("category does not belong to user") // Категория принадлежит другому пользователю
		}
	} else {
		// Если CategoryID пуст, это может быть нежелательно,
		// в зависимости от бизнес-логики.
		// Можно вернуть ошибку или присвоить "дефолтную" категорию.
		// Сейчас просто пропускаем, если не указано, но обычно CategoryID required.
		return errors.New("category ID is required for a transaction") // Добавим требование CategoryID
	}

	input.UserID = userID
	input.CreatedAt = time.Now()
	// GORM автоматически установит ID, если поле ID в модели string и помечено default:gen_random_uuid()
	// input.ID = uuid.New().String() // Это больше не нужно, если GORM сам генерирует UUID

	return s.repo.Create(input)
}

// List получает список транзакций для пользователя.
func (s *txService) List(userID string, from, to *time.Time, txType string) ([]model.Transaction, error) {
	return s.repo.GetByUser(userID, from, to, txType)
}

// Update обновляет существующую транзакцию.
func (s *txService) Update(id, userID string, input *model.Transaction) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("transaction not found") // Транзакция не найдена
	}
	if existing.UserID != userID {
		return errors.New("transaction does not belong to user") // Транзакция принадлежит другому пользователю
	}

	// Валидация CategoryID для обновления
	if input.CategoryID != "" {
		category, err := s.categoryRepo.GetByID(input.CategoryID)
		if err != nil {
			return errors.New("category not found") // Категория не найдена
		}
		if category.UserID != userID {
			return errors.New("category does not belong to user") // Категория принадлежит другому пользователю
		}
	} else {
		return errors.New("category ID is required for a transaction update") // Добавим требование CategoryID
	}

	existing.Amount = input.Amount
	existing.Type = input.Type
	existing.CategoryID = input.CategoryID // <-- НОВОЕ: Обновляем CategoryID
	// existing.Category = input.Category // Эту строку нужно удалить, так как Category - это отношение, а не поле для обновления
	existing.Comment = input.Comment // <-- ИСПРАВЛЕНО: Теперь Comment стал Description
	existing.Date = input.Date       // <-- НОВОЕ: Обновляем поле Date

	return s.repo.Update(existing)
}

// Delete удаляет транзакцию.
func (s *txService) Delete(id, userID string) error {
	tx, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("transaction not found")
	}
	if tx.UserID != userID {
		return errors.New("transaction does not belong to user")
	}
	return s.repo.Delete(id)
}

// Summary подсчитывает общую сумму доходов и расходов.
func (s *txService) Summary(userID string, from, to *time.Time) (float64, float64, error) {
	return s.repo.Summary(userID, from, to)
}

// ByCategory группирует транзакции по категориям.
func (s *txService) ByCategory(userID string, from, to *time.Time) (map[string]float64, error) {
	return s.repo.ByCategory(userID, from, to)
}
