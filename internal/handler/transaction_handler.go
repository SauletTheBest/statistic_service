package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/model"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct{ svc service.TransactionService }

func NewTransactionHandler(s service.TransactionService) *TransactionHandler {
	return &TransactionHandler{s}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var input model.Transaction
	if c.BindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	userID := c.GetString("userID")
	if err := h.svc.Create(userID, &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

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
	list, _ := h.svc.List(userID, from, to, txType)
	c.JSON(http.StatusOK, list)
}

func (h *TransactionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	var input model.Transaction
	if c.BindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	if err := h.svc.Update(id, userID, &input); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	if err := h.svc.Delete(id, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
