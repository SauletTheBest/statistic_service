// internal/handler/transaction_handler.go
package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/model"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TransactionHandler struct {
	svc    service.TransactionService
	logger *logrus.Logger
}

func NewTransactionHandler(s service.TransactionService, logger *logrus.Logger) *TransactionHandler {
	return &TransactionHandler{svc: s, logger: logger}
}

// Create godoc
// @Summary Create a new transaction
// @Description Adds a new transaction (expense or income) for the authenticated user
// @Tags Transactions
// @Accept json
// @Produce json
// @Param transaction body model.Transaction true "Transaction details"
// @Success 201 "Created"
// @Failure 400 {object} map[string]string "error: bad request"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	var input model.Transaction
	if err := c.BindJSON(&input); err != nil {
		h.logger.WithError(err).Warn("Invalid transaction payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	userID := c.GetString("userID")
	h.logger.WithFields(logrus.Fields{"userID": userID, "amount": input.Amount, "type": input.Type}).Info("Creating transaction")
	if err := h.svc.Create(userID, &input); err != nil {
		h.logger.WithError(err).Error("Failed to create transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Security BearerAuth
// @Router /transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	userID := c.GetString("userID")
	var from, to *time.Time
	if f := c.Query("date_from"); f != "" {
		t, _ := time.Parse(time.RFC3339, f)
		from = &t
	}
	if tstr := c.Query("date_to"); tstr != "" {
		t, _ := time.Parse(time.RFC3339, tstr)
		to = &t
	}
	txType := c.Query("type")
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to, "type": txType}).Info("Listing transactions")
	list, err := h.svc.List(userID, from, to, txType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list transactions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logger.WithField("count", len(list)).Info("Transactions listed successfully")
	c.JSON(http.StatusOK, list)
}

// Update godoc
// @Summary Update a transaction
// @Description Updates an existing transaction by ID for the authenticated user
// @Tags Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param transaction body model.Transaction true "Updated transaction details"
// @Success 200 "OK"
// @Failure 400 {object} map[string]string "error: bad request"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 403 {object} map[string]string "error: forbidden"
// @Security BearerAuth
// @Router /transactions/{id} [put]
func (h *TransactionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	var input model.Transaction
	if err := c.BindJSON(&input); err != nil {
		h.logger.WithError(err).Warn("Invalid update payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	h.logger.WithFields(logrus.Fields{"transactionID": id, "userID": userID}).Info("Updating transaction")
	if err := h.svc.Update(id, userID, &input); err != nil {
		h.logger.WithError(err).Error("Failed to update transaction")
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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
// @Failure 403 {object} map[string]string "error: forbidden"
// @Security BearerAuth
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	h.logger.WithFields(logrus.Fields{"transactionID": id, "userID": userID}).Info("Deleting transaction")
	if err := h.svc.Delete(id, userID); err != nil {
		h.logger.WithError(err).Error("Failed to delete transaction")
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("Transaction deleted successfully")
	c.Status(http.StatusNoContent)
}
