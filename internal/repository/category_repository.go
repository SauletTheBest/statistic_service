package repository

import (
	"statistic_service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository определяет интерфейс для операций с категориями в базе данных.
type CategoryRepository interface {
	CreateCategory(category *model.Category) error
	GetByID(id string) (model.Category, error)
	GetCategoryByID(id uuid.UUID, userID uuid.UUID) (*model.Category, error)
	GetCategoriesByUserID(userID uuid.UUID, categoryType string) ([]model.Category, error)
	UpdateCategory(category *model.Category) error
	DeleteCategory(id uuid.UUID, userID uuid.UUID) error
}

// categoryRepository реализует CategoryRepository, используя GORM.
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository создает новый экземпляр CategoryRepository.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// CreateCategory создает новую категорию в базе данных.
func (r *categoryRepository) CreateCategory(category *model.Category) error {
	return r.db.Create(category).Error
}

// GetByID реализует получение категории по её ID.
func (r *categoryRepository) GetByID(id string) (model.Category, error) {
	var category model.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		return model.Category{}, err
	}
	return category, nil
}

// GetCategoryByID получает категорию по её ID и ID пользователя.
func (r *categoryRepository) GetCategoryByID(id uuid.UUID, userID uuid.UUID) (*model.Category, error) {
	var category model.Category
	// Важно убедиться, что пользователь может получить только свои категории
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoriesByUserID получает все категории для конкретного пользователя.
// categoryType может быть "income", "expense" или пустым для всех типов.
func (r *categoryRepository) GetCategoriesByUserID(userID uuid.UUID, categoryType string) ([]model.Category, error) {
	var categories []model.Category
	query := r.db.Where("user_id = ?", userID)
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}
	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// UpdateCategory обновляет существующую категорию в базе данных.
// Важно: Эта функция обновляет поля на основе переданной модели,
// но перед вызовом убедитесь, что категория принадлежит пользователю,
// вызвав GetCategoryByID или аналогичный метод.
func (r *categoryRepository) UpdateCategory(category *model.Category) error {
	// GORM по умолчанию обновляет только непустые поля.
	// Если нужно обновить все поля, включая пустые строки или 0, используйте r.db.Save(category)
	// или r.db.Model(&category).Updates(...)
	return r.db.Save(category).Error
}

// DeleteCategory удаляет категорию по её ID и ID пользователя.
func (r *categoryRepository) DeleteCategory(id uuid.UUID, userID uuid.UUID) error {
	// Удаляем только категории, которые принадлежат указанному пользователю
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Если ни одна запись не была удалена, это означает, что категория не найдена или не принадлежит пользователю
	}
	return nil
}
