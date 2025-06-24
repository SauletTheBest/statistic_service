package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/model"
	"statistic_service/internal/service"
	"statistic_service/pkg/utils"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type TransactionHandler struct {
	svc    service.TransactionService
	logger *logrus.Logger
}

func NewTransactionHandler(s service.TransactionService, logger *logrus.Logger) *TransactionHandler {
	return &TransactionHandler{svc: s, logger: logger}
}

// CreateTransactionRequest представляет тело запроса для создания транзакции.
type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // RFC3339 format
	CategoryID  string  `json:"category_id" binding:"required,uuid"`                        // <-- НОВОЕ: Требуется CategoryID и формат UUID
}

// UpdateTransactionRequest представляет тело запроса для обновления транзакции.
type UpdateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // RFC3339 format
	CategoryID  string  `json:"category_id" binding:"required,uuid"`                        // <-- НОВОЕ: Требуется CategoryID и формат UUID
}

// Create godoc
// @Summary Create a new transaction
// @Description Adds a new transaction (expense or income) for the authenticated user, linking to a category.
// @Tags Transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction details including category ID"
// @Success 201 "Created"
// @Failure 400 {object} map[string]string "error: bad request (e.g., validation failed, invalid category ID)"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid transaction payload")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.SendErrorResponse(c, http.StatusBadRequest, strings.Join(utils.CustomValidationErrors(validationErrors), ", "))
		} else {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		}
		return
	}

	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid date format in transaction payload")
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid date format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
		return
	}

	userID := c.GetString("userID")
	h.logger.WithFields(logrus.Fields{
		"userID":     userID,
		"amount":     req.Amount,
		"type":       req.Type,
		"categoryID": req.CategoryID,
	}).Info("Creating transaction")

	transaction := &model.Transaction{
		Amount:     req.Amount,
		Type:       req.Type,
		Comment:    req.Description,
		Date:       parsedDate,
		CategoryID: req.CategoryID, // <-- НОВОЕ: Присваиваем CategoryID из запроса
	}

	if err := h.svc.Create(userID, transaction); err != nil {
		h.logger.WithError(err).Error("Failed to create transaction")
		if strings.Contains(err.Error(), "category not found") || strings.Contains(err.Error(), "category does not belong to user") || strings.Contains(err.Error(), "category ID is required") {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.logger.Info("Transaction created successfully")
	c.Status(http.StatusCreated)
}

// List godoc
// @Summary List transactions
// @Description Retrieves transactions for the authenticated user, with optional filters
// @Tags Transactions
// @Accept json
// @Produce json
// @Param date_from query string false "Start date in RFC3339 format"
// @Param date_to query string false "End date in RFC3339 format"
// @Param type query string false "Transaction type: income or expense"
// @Success 200 {array} model.Transaction
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	userID := c.GetString("userID")
	var from, to *time.Time
	if f := c.Query("date_from"); f != "" {
		t, err := time.Parse(time.RFC3339, f)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_from format")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_from' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		from = &t
	}
	if tstr := c.Query("date_to"); tstr != "" {
		t, err := time.Parse(time.RFC3339, tstr)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_to format")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_to' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		to = &t
	}
	txType := c.Query("type")
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to, "type": txType}).Info("Listing transactions")
	list, err := h.svc.List(userID, from, to, txType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list transactions")
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.logger.WithField("count", len(list)).Info("Transactions listed successfully")
	c.JSON(http.StatusOK, list)
}

// Update godoc
// @Summary Update a transaction
// @Description Updates an existing transaction by ID for the authenticated user, including category.
// @Tags Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param transaction body UpdateTransactionRequest true "Updated transaction details including category ID"
// @Success 200 "OK"
// @Failure 400 {object} map[string]string "error: bad request (e.g., validation failed, invalid category ID)"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 403 {object} map[string]string "error: forbidden (e.g., transaction not found or not owned by user)"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /transactions/{id} [put]
func (h *TransactionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update payload")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.SendErrorResponse(c, http.StatusBadRequest, strings.Join(utils.CustomValidationErrors(validationErrors), ", "))
		} else {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		}
		return
	}

	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid date format in update payload")
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid date format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
		return
	}

	h.logger.WithFields(logrus.Fields{"transactionID": id, "userID": userID}).Info("Updating transaction")

	transaction := &model.Transaction{
		Amount:     req.Amount,
		Type:       req.Type,
		Comment:    req.Description,
		Date:       parsedDate,
		CategoryID: req.CategoryID, // <-- НОВОЕ: Присваиваем CategoryID из запроса
	}

	if err := h.svc.Update(id, userID, transaction); err != nil {
		h.logger.WithError(err).Error("Failed to update transaction")
		// Более детальная обработка ошибок из сервиса
		if strings.Contains(err.Error(), "transaction not found") || strings.Contains(err.Error(), "transaction does not belong to user") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error()) // 403 Forbidden если не принадлежит или не найдена
		} else if strings.Contains(err.Error(), "category not found") || strings.Contains(err.Error(), "category does not belong to user") || strings.Contains(err.Error(), "category ID is required") {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error()) // 400 Bad Request для ошибок категории
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.logger.Info("Transaction updated successfully")
	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a transaction
// @Description Deletes an existing transaction by ID for the authenticated user
// @Tags Transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 403 {object} map[string]string "error: forbidden (e.g., transaction not found or not owned by user)"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	h.logger.WithFields(logrus.Fields{"transactionID": id, "userID": userID}).Info("Deleting transaction")
	if err := h.svc.Delete(id, userID); err != nil {
		h.logger.WithError(err).Error("Failed to delete transaction")
		if strings.Contains(err.Error(), "transaction not found") || strings.Contains(err.Error(), "transaction does not belong to user") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error()) // 403 Forbidden если не принадлежит или не найдена
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.logger.Info("Transaction deleted successfully")
	c.Status(http.StatusNoContent)
}
