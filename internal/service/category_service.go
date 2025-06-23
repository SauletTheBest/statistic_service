package service

import (
	"errors"
	"statistic_service/internal/logger"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryService определяет интерфейс для бизнес-логики категорий.
type CategoryService interface {
	CreateCategory(userID uuid.UUID, name, categoryType string) (*model.Category, error)
	GetCategoryByID(id uuid.UUID, userID uuid.UUID) (*model.Category, error)
	ListCategories(userID uuid.UUID, categoryType string) ([]model.Category, error)
	UpdateCategory(categoryID uuid.UUID, userID uuid.UUID, newName, newType string) (*model.Category, error)
	DeleteCategory(id uuid.UUID, userID uuid.UUID) error
}

// categoryService реализует CategoryService.
type categoryService struct {
	categoryRepo repository.CategoryRepository
	logger       *logger.Logger
}

// NewCategoryService создает новый экземпляр CategoryService.
func NewCategoryService(categoryRepo repository.CategoryRepository, logger *logger.Logger) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// CreateCategory создает новую категорию для указанного пользователя.
func (s *categoryService) CreateCategory(userID uuid.UUID, name, categoryType string) (*model.Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if categoryType != "income" && categoryType != "expense" {
		return nil, errors.New("category type must be 'income' or 'expense'")
	}

	category := &model.Category{
		ID:     uuid.New(),
		UserID: userID,
		Name:   name,
		Type:   categoryType,
	}

	if err := s.categoryRepo.CreateCategory(category); err != nil {
		s.logger.Errorf("Failed to create category for user %s: %v", userID, err)
		return nil, errors.New("failed to create category")
	}
	s.logger.Infof("Category '%s' created successfully for user %s", name, userID)
	return category, nil
}

// GetCategoryByID получает категорию по её ID.
func (s *categoryService) GetCategoryByID(id uuid.UUID, userID uuid.UUID) (*model.Category, error) {
	category, err := s.categoryRepo.GetCategoryByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Category with ID %s not found for user %s", id, userID)
			return nil, errors.New("category not found or not accessible")
		}
		s.logger.Errorf("Failed to get category %s for user %s: %v", id, userID, err)
		return nil, errors.New("failed to retrieve category")
	}
	return category, nil
}

// ListCategories получает список всех категорий для указанного пользователя.
func (s *categoryService) ListCategories(userID uuid.UUID, categoryType string) ([]model.Category, error) {
	if categoryType != "" && categoryType != "income" && categoryType != "expense" {
		return nil, errors.New("invalid category type specified, must be 'income', 'expense', or empty")
	}

	categories, err := s.categoryRepo.GetCategoriesByUserID(userID, categoryType)
	if err != nil {
		s.logger.Errorf("Failed to list categories for user %s: %v", userID, err)
		return nil, errors.New("failed to retrieve categories")
	}
	return categories, nil
}

// UpdateCategory обновляет существующую категорию.
func (s *categoryService) UpdateCategory(categoryID uuid.UUID, userID uuid.UUID, newName, newType string) (*model.Category, error) {
	if newName == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if newType != "income" && newType != "expense" {
		return nil, errors.New("category type must be 'income' or 'expense'")
	}

	existingCategory, err := s.categoryRepo.GetCategoryByID(categoryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Category with ID %s not found for user %s during update attempt", categoryID, userID)
			return nil, errors.New("category not found or not accessible")
		}
		s.logger.Errorf("Failed to retrieve category %s for user %s during update: %v", categoryID, userID, err)
		return nil, errors.New("failed to update category")
	}

	existingCategory.Name = newName
	existingCategory.Type = newType

	if err := s.categoryRepo.UpdateCategory(existingCategory); err != nil {
		s.logger.Errorf("Failed to update category %s for user %s: %v", categoryID, userID, err)
		return nil, errors.New("failed to update category")
	}
	s.logger.Infof("Category %s updated successfully for user %s", categoryID, userID)
	return existingCategory, nil
}

// DeleteCategory удаляет категорию.
func (s *categoryService) DeleteCategory(id uuid.UUID, userID uuid.UUID) error {
	err := s.categoryRepo.DeleteCategory(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Category with ID %s not found for user %s during deletion attempt", id, userID)
			return errors.New("category not found or not accessible")
		}
		s.logger.Errorf("Failed to delete category %s for user %s: %v", id, userID, err)
		return errors.New("failed to delete category")
	}
	s.logger.Infof("Category %s deleted successfully for user %s", id, userID)
	return nil
}
