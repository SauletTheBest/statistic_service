package handler

import (
	"errors"
	"net/http"
	"statistic_service/internal/service"
	"statistic_service/pkg/utils" // Убедись, что это правильный путь к файлу utils/errors.go
	"strings"                     // <-- НОВОЕ: Добавляем импорт пакета strings

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // <-- НОВОЕ: Добавляем импорт для validator.ValidationErrors
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CategoryHandler определяет структуру для хендлеров категорий.
type CategoryHandler struct {
	categoryService service.CategoryService
	logger          *logrus.Logger
}

// NewCategoryHandler создает новый экземпляр CategoryHandler.
func NewCategoryHandler(categoryService service.CategoryService, logger *logrus.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// CreateCategoryRequest представляет тело запроса для создания категории.
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

// CreateCategory @Summary Создать новую категорию
// @Description Создает новую категорию для текущего пользователя.
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param category body CreateCategoryRequest true "Данные для создания категории"
// @Success 201 {object} model.Category
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("UserID not found in context for CreateCategory")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Errorf("Failed to parse UserID from context: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Invalid request body for CreateCategory: %v", err)
		// Исправлено: используем CustomValidationErrors и объединяем результат
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.SendErrorResponse(c, http.StatusBadRequest, strings.Join(utils.CustomValidationErrors(validationErrors), ", "))
		} else {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	category, err := h.categoryService.CreateCategory(parsedUserID, req.Name, req.Type)
	if err != nil {
		h.logger.Errorf("Error creating category for user %s: %v", parsedUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetCategoryByID @Summary Получить категорию по ID
// @Description Получает детали категории по её ID для текущего пользователя.
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param id path string true "ID категории"
// @Success 200 {object} model.Category
// @Failure 400 {object} map[string]string "Неверный ID категории"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		h.logger.Warnf("Invalid category ID format: %s", categoryIDStr)
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("UserID not found in context for GetCategoryByID")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Errorf("Failed to parse UserID from context: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	category, err := h.categoryService.GetCategoryByID(categoryID, parsedUserID)
	if err != nil {
		if errors.Is(err, errors.New("category not found or not accessible")) {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Errorf("Error getting category %s for user %s: %v", categoryID, parsedUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, category)
}

// ListCategories @Summary Получить все категории
// @Description Получает список всех категорий для текущего пользователя, опционально фильтруя по типу.
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param type query string false "Тип категории (income или expense)"
// @Success 200 {array} model.Category
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("UserID not found in context for ListCategories")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Errorf("Failed to parse UserID from context: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	categoryType := c.Query("type") // Получаем параметр 'type' из запроса

	categories, err := h.categoryService.ListCategories(parsedUserID, categoryType)
	if err != nil {
		h.logger.Errorf("Error listing categories for user %s: %v", parsedUserID, err)
		if errors.Is(err, errors.New("invalid category type specified, must be 'income', 'expense', or empty")) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, categories)
}

// UpdateCategoryRequest представляет тело запроса для обновления категории.
type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

// UpdateCategory @Summary Обновить категорию
// @Description Обновляет существующую категорию по её ID для текущего пользователя.
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param id path string true "ID категории"
// @Param category body UpdateCategoryRequest true "Обновленные данные категории"
// @Success 200 {object} model.Category
// @Failure 400 {object} map[string]string "Неверный ID категории или данные запроса"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		h.logger.Warnf("Invalid category ID format for update: %s", categoryIDStr)
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("UserID not found in context for UpdateCategory")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Errorf("Failed to parse UserID from context: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Invalid request body for UpdateCategory: %v", err)
		// Исправлено: используем CustomValidationErrors и объединяем результат
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.SendErrorResponse(c, http.StatusBadRequest, strings.Join(utils.CustomValidationErrors(validationErrors), ", "))
		} else {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	updatedCategory, err := h.categoryService.UpdateCategory(categoryID, parsedUserID, req.Name, req.Type)
	if err != nil {
		if errors.Is(err, errors.New("category not found or not accessible")) {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Errorf("Error updating category %s for user %s: %v", categoryID, parsedUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory @Summary Удалить категорию
// @Description Удаляет категорию по её ID для текущего пользователя.
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param id path string true "ID категории"
// @Success 204 "Категория успешно удалена"
// @Failure 400 {object} map[string]string "Неверный ID категории"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 404 {object} map[string]string "Категория не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		h.logger.Warnf("Invalid category ID format for delete: %s", categoryIDStr)
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("UserID not found in context for DeleteCategory")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Errorf("Failed to parse UserID from context: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	err = h.categoryService.DeleteCategory(categoryID, parsedUserID)
	if err != nil {
		if errors.Is(err, errors.New("category not found or not accessible")) {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Errorf("Error deleting category %s for user %s: %v", categoryID, parsedUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}
