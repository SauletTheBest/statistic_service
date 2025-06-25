package handler

import (
	"errors"
	"net/http"
	"statistic_service/internal/model"
	"statistic_service/internal/service"
	"statistic_service/pkg/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type WalletHandler struct {
	walletService service.WalletService
	logger        *logrus.Logger
}

func NewWalletHandler(ws service.WalletService, logger *logrus.Logger) *WalletHandler {
	return &WalletHandler{walletService: ws, logger: logger}
}

type createWalletRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateWalletRequest struct {
	Name string `json:"name" binding:"required"`
}

type inviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

type CreateTransactionRequest struct {
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Type       string  `json:"type" binding:"required,oneof=income expense"`
	Comment    string  `json:"comment"`
	Date       string  `json:"date" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // RFC3339
	CategoryID string  `json:"category_id" binding:"required,uuid"`
}

// Create @Summary Создать новый кошелек
// ... (добавьте аннотации для Swagger)
func (h *WalletHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	var req createWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	wallet, err := h.walletService.CreateWallet(userID, req.Name)
	if err != nil {
		// Здесь можно проверять тип ошибки от сервиса и возвращать разные статусы
		// Например, 409 Conflict если "wallet limit reached"
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// List @Summary Получить список кошельков пользователя
// @Description Возвращает все кошельки, в которых состоит текущий пользователь
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Wallet "Список кошельков"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /wallets [get]
func (h *WalletHandler) List(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	wallets, err := h.walletService.ListWallets(userID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Если кошельков нет, вернем пустой массив, а не null
	if wallets == nil {
		wallets = []model.Wallet{}
	}

	c.JSON(http.StatusOK, wallets)
}

// --- Замените заглушку GetByID ---
// GetByID @Summary Получить кошелек по ID
// @Description Получает детали одного кошелька, если пользователь является его участником
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Success 200 {object} model.Wallet
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Кошелек не найден или доступ запрещен"
// @Router /wallets/{id} [get]
func (h *WalletHandler) GetByID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	wallet, err := h.walletService.GetWalletByID(walletID, userID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// --- Замените заглушку UpdateName ---
// UpdateName @Summary Обновить имя кошелька
// @Description Обновляет имя кошелька. Требуются права администратора кошелька.
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Param wallet body updateWalletRequest true "Новое имя кошелька"
// @Success 200 {object} model.Wallet
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Кошелек не найден"
// @Router /wallets/{id} [put]
func (h *WalletHandler) UpdateName(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	var req updateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: name is required")
		return
	}

	updatedWallet, err := h.walletService.UpdateWalletName(walletID, userID, req.Name)
	if err != nil {
		// Определяем тип ошибки для правильного HTTP статуса
		if strings.Contains(err.Error(), "permission denied") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, updatedWallet)
}

// --- Замените заглушку Delete ---
// Delete @Summary Удалить кошелек
// @Description Удаляет кошелек. Требуются права владельца кошелька.
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Success 204 "No Content"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Кошелек не найден"
// @Router /wallets/{id} [delete]
func (h *WalletHandler) Delete(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	err = h.walletService.DeleteWallet(walletID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// InviteMember @Summary Пригласить пользователя в кошелек
// @Description Приглашает пользователя по email. Требуются права администратора кошелька.
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Param invitation body inviteMemberRequest true "Email пользователя для приглашения"
// @Success 201 "Created"
// @Failure 400 {object} map[string]string "Неверный email"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Пользователь с таким email не найден"
// @Failure 409 {object} map[string]string "Пользователь уже является участником"
// @Router /wallets/{id}/invite [post]
func (h *WalletHandler) InviteMember(c *gin.Context) {
	inviterID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	var req inviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: valid email is required")
		return
	}

	err = h.walletService.InviteUserToWallet(walletID, inviterID, req.Email)
	if err != nil {
		// Более детальная обработка ошибок от сервиса
		if strings.Contains(err.Error(), "permission denied") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "not found") {
			utils.SendErrorResponse(c, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already a member") {
			utils.SendErrorResponse(c, http.StatusConflict, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.Status(http.StatusCreated)
}

// --- Замените заглушку GetMembers ---
// GetMembers @Summary Получить список участников кошелька
// @Description Возвращает всех участников кошелька. Доступно любому участнику.
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Success 200 {array} model.WalletMember "Список участников"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Кошелек не найден"
// @Router /wallets/{id}/members [get]
func (h *WalletHandler) GetMembers(c *gin.Context) {
	requesterID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	members, err := h.walletService.GetWalletMembers(walletID, requesterID)
	if err != nil {
		// Ошибка "not found or access denied" может означать и 404, и 403
		utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	if members == nil {
		members = []model.WalletMember{}
	}

	c.JSON(http.StatusOK, members)
}

// UpdateMemberRole @Summary Изменить роль участника
// @Description Изменяет роль участника на 'admin' или 'member'. Требуются права администратора.
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Param userID path string true "ID Участника, чью роль меняют"
// @Param role body updateRoleRequest true "Новая роль"
// @Success 200 "OK"
// @Failure 400 {object} map[string]string "Неверный формат ID или данных запроса"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Router /wallets/{id}/members/{userID}/role [put]
func (h *WalletHandler) UpdateMemberRole(c *gin.Context) {
	adminID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	memberID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid member user ID format")
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: role must be 'admin' or 'member'")
		return
	}

	err = h.walletService.UpdateMemberRole(walletID, adminID, memberID, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "cannot change") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "role updated successfully"})
}

// --- Замените заглушку RemoveMember ---
// RemoveMember @Summary Удалить участника из кошелька
// @Description Удаляет участника из кошелька. Требуются права администратора.
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Param userID path string true "ID Участника для удаления"
// @Success 204 "No Content"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Router /wallets/{id}/members/{userID} [delete]
func (h *WalletHandler) RemoveMember(c *gin.Context) {
	removerID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	memberToRemoveID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid member user ID format")
		return
	}

	err = h.walletService.RemoveMemberFromWallet(walletID, removerID, memberToRemoveID)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "cannot remove") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Используем ту же структуру запроса, что и для обычных транзакций.
// Поле WalletID в теле будет проигнорировано, так как ID берется из URL.

// --- Замените заглушку CreateTransaction ---
// CreateTransaction @Summary Создать транзакцию в кошельке
// @Description Создает новую транзакцию в указанном кошельке.
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Param transaction body CreateTransactionRequest true "Данные транзакции"
// @Success 201 {object} model.Transaction
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Router /wallets/{id}/transactions [post]
func (h *WalletHandler) CreateTransaction(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Можно добавить более детальную обработку ошибок валидации как в transaction_handler
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid date format. Use RFC3339")
		return
	}

	tx := &model.Transaction{
		Amount:     req.Amount,
		Type:       req.Type,
		Comment:    req.Comment,
		Date:       parsedDate,
		CategoryID: req.CategoryID,
	}

	createdTx, err := h.walletService.CreateTransactionInWallet(walletID, userID, tx)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, createdTx)
}

// --- Замените заглушку GetTransactions ---
// GetTransactions @Summary Получить транзакции кошелька
// @Description Возвращает список всех транзакций для указанного кошелька.
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID Кошелька"
// @Success 200 {array} model.Transaction
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Router /wallets/{id}/transactions [get]
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID format")
		return
	}

	transactions, err := h.walletService.GetTransactionsForWallet(walletID, userID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	c.JSON(http.StatusOK, transactions)
}

// Вспомогательная функция для получения userID
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format in context")
	}
	return userID, nil
}
