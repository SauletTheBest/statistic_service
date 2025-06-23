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
	return r.db.Create(tx).Error
}

func (r *transactionRepository) GetByUser(userID string, from, to *time.Time, txType string) ([]model.Transaction, error) {
	q := r.db.Where("user_id = ?", userID)
	if txType != "" {
		q = q.Where("type = ?", txType)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	var transactions []model.Transaction
	if err := q.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByID(id string) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.db.First(&tx, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) Update(tx *model.Transaction) error {
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
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
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
	rows, err := r.db.Raw(
		"SELECT category, SUM(amount) FROM transactions WHERE user_id = ? AND created_at BETWEEN ? AND ? GROUP BY category",
		userID, from, to,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var sum float64
		rows.Scan(&category, &sum)
		results[category] = sum
	}
	return results, nil
}
