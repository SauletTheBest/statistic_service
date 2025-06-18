package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/service"

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
// @Failure 400 {object} map[string]string "Invalid range"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /stats/timeline [get]
func (h *TimelineHandler) Timeline(c *gin.Context) {
	userID := c.GetString("userID")
	txType := c.DefaultQuery("type", "expense") // default to expense
	rangeType := c.DefaultQuery("range", "month")

	now := time.Now()
	var startDate time.Time

	switch rangeType {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid range"})
		return
	}

	list, err := h.svc.List(userID, &startDate, &now, txType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get timeline transactions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make(map[string]float64)
	for d := startDate; d.Before(now) || d.Equal(now); d = d.AddDate(0, 0, 1) {
		result[d.Format("2006-01-02")] = 0
	}

	for _, tx := range list {
		date := tx.CreatedAt.Format("2006-01-02")
		result[date] += tx.Amount
	}

	c.JSON(http.StatusOK, result)
}
