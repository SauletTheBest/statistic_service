// internal/handler/wallet_handler.go
package handler

import (
	"net/http"
	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	service *service.WalletService
}

func NewWalletHandler(service *service.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

func (h *WalletHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	wallet, err := h.service.CreateWallet(userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

func (h *WalletHandler) Invite(c *gin.Context) {
	walletID := uuid.MustParse(c.Param("id"))
	var req struct {
		UserID uuid.UUID `json:"user_id"`
		Role   string    `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.InviteMember(walletID, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	walletID := uuid.MustParse(c.Param("id"))
	txs, err := h.service.GetTransactions(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}

func (h *WalletHandler) GetMembers(c *gin.Context) {
	walletID := uuid.MustParse(c.Param("id"))
	members, err := h.service.GetMembers(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *WalletHandler) CreateTransaction(c *gin.Context) {
	walletID := uuid.MustParse(c.Param("id"))

	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Amount     float64    `json:"amount"`
		CategoryID *uuid.UUID `json:"category_id"`
		Comment    string     `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.service.CreateTransaction(walletID, userID, req.Amount, req.CategoryID, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tx)
}
