package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/service"
	// "statistic_service/internal/utils" // Пока не требуется, если нет обработки ошибок парсинга даты

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TimelineHandler struct {
	svc    service.TransactionService
	logger *logrus.Logger
}

func NewTimelineHandler(s service.TransactionService, logger *logrus.Logger) *TimelineHandler {
	return &TimelineHandler{svc: s, logger: logger}
}

// Timeline godoc
// @Summary Get timeline of expenses or income
// @Description Returns a map of daily totals over a specified time range (week or month) for graph/chart usage
// @Tags Statistics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Transaction type: expense or income" default(expense)
// @Param range query string false "Time range: week or month" default(month)
// @Success 200 {object} map[string]float64
// @Failure 400 {object} map[string]string "Invalid range or type"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /stats/timeline [get]
func (h *TimelineHandler) Timeline(c *gin.Context) {
	userID := c.GetString("userID")
	txType := c.DefaultQuery("type", "expense") // default to expense
	rangeType := c.DefaultQuery("range", "month")

	// Валидация txType
	if txType != "expense" && txType != "income" {
		h.logger.WithField("txType", txType).Warn("Invalid transaction type for timeline")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type. Must be 'expense' or 'income'"})
		return
	}

	now := time.Now()
	var startDate time.Time

	switch rangeType {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		h.logger.WithField("rangeType", rangeType).Warn("Invalid range type for timeline")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid range. Must be 'week' or 'month'"})
		return
	}

	// Получаем список транзакций за указанный период (используя поле Date в TransactionRepository)
	list, err := h.svc.List(userID, &startDate, &now, txType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get timeline transactions from service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Инициализируем map для всех дней в диапазоне с 0
	result := make(map[string]float64)
	// Важно: здесь теперь включаем и конечную дату (now), чтобы она тоже была в ключах,
	// так как транзакции могут быть за сегодня.
	// Или более точно: от startDate до today.
	for d := startDate; !d.After(now); d = d.AddDate(0, 0, 1) {
		result[d.Format("2006-01-02")] = 0
	}

	// Суммируем транзакции по их фактической дате (tx.Date)
	for _, tx := range list {
		// <-- ИСПРАВЛЕНО: Используем tx.Date вместо tx.CreatedAt
		dateStr := tx.Date.Format("2006-01-02")
		result[dateStr] += tx.Amount
	}

	h.logger.Info("Timeline data retrieved successfully")
	c.JSON(http.StatusOK, result)
}
