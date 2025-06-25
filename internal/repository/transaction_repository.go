package repository

import (
	"time"

	"statistic_service/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(tx *model.Transaction) error
	GetByUser(userID string, from, to *time.Time, txType string) ([]model.Transaction, error)
	GetByID(id string) (*model.Transaction, error)
	Update(tx *model.Transaction) error
	Delete(id string) error
	Summary(userID string, from, to *time.Time) (income, expense float64, err error)
	ByCategory(userID string, from, to *time.Time) (map[string]float64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *model.Transaction) error {
	// GORM автоматически сохранит CategoryID, так как оно теперь поле модели Transaction
	return r.db.Create(tx).Error
}

func (r *transactionRepository) GetByUser(userID string, from, to *time.Time, txType string) ([]model.Transaction, error) {
	var transactions []model.Transaction
	q := r.db.Preload("Category").Where("user_id = ?", userID) // <-- НОВОЕ: Preload("Category")
	if txType != "" {
		q = q.Where("type = ?", txType)
	}
	if from != nil {
		q = q.Where("date >= ?", *from) // Изменено на 'date' если это поле для фильтрации по дате
	}
	if to != nil {
		q = q.Where("date <= ?", *to) // Изменено на 'date' если это поле для фильтрации по дате
	}
	// Если ты хочешь фильтровать по CreatedAt, верни `created_at`
	// Если `date` - это фактическая дата транзакции, то лучше использовать её.
	if err := q.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByID(id string) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.db.Preload("Category").First(&tx, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) Update(tx *model.Transaction) error {
	// GORM Save обновляет все поля, включая CategoryID
	return r.db.Save(tx).Error
}

func (r *transactionRepository) Delete(id string) error {
	return r.db.Delete(&model.Transaction{}, "id = ?", id).Error
}

func (r *transactionRepository) Summary(userID string, from, to *time.Time) (float64, float64, error) {
	var income, expense float64
	type row struct {
		Type string
		Sum  float64
	}
	var rows []row

	q := r.db.Model(&model.Transaction{}).
		Select("type, SUM(amount) as sum").
		Where("user_id = ?", userID)

	if from != nil {
		q = q.Where("date >= ?", *from) // Изменено на 'date'
	}
	if to != nil {
		q = q.Where("date <= ?", *to) // Изменено на 'date'
	}

	if err := q.Group("type").Scan(&rows).Error; err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		if r.Type == "income" {
			income = r.Sum
		} else if r.Type == "expense" {
			expense = r.Sum
		}
	}
	return income, expense, nil
}

func (r *transactionRepository) ByCategory(userID string, from, to *time.Time) (map[string]float64, error) {
	results := make(map[string]float64)
	type CategorySum struct {
		CategoryName string  `gorm:"column:category_name"` // Имя столбца для результата
		TotalAmount  float64 `gorm:"column:total_amount"`
	}
	var categorySums []CategorySum

	query := r.db.Model(&model.Transaction{}).
		Select("categories.name AS category_name, SUM(transactions.amount) AS total_amount").
		Joins("JOIN categories ON transactions.category_id = categories.id"). // <-- НОВОЕ: INNER JOIN
		Where("transactions.user_id = ?", userID)

	if from != nil {
		query = query.Where("transactions.date >= ?", *from) // Изменено на 'date'
	}
	if to != nil {
		query = query.Where("transactions.date <= ?", *to) // Изменено на 'date'
	}

	if err := query.Group("categories.name").Scan(&categorySums).Error; err != nil { // <-- НОВОЕ: Группировка по categories.name
		return nil, err
	}

	for _, cs := range categorySums {
		results[cs.CategoryName] = cs.TotalAmount
	}
	return results, nil
}
