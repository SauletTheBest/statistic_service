package handler

import (
	"net/http"
	"statistic_service/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PredictHandler struct {
	svc    service.TransactionService
	logger *logrus.Logger
}

func NewPredictHandler(s service.TransactionService, logger *logrus.Logger) *PredictHandler {
	return &PredictHandler{svc: s, logger: logger}
}

// Predict godoc
// @Summary Predict next month's expenses or income
// @Description Calculates the expected total expenses or income for the next month based on the current month's average
// @Tags Statistics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string true "Transaction type: expense or income"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} map[string]string "Invalid type"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /predict [get]
func (h *PredictHandler) Predict(c *gin.Context) {
	userID := c.GetString("userID")
	txType := c.Query("type")
	if txType != "expense" && txType != "income" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := currentMonthStart.AddDate(0, 1, 0)
	nextMonthEnd := nextMonthStart.AddDate(0, 1, -1)

	list, err := h.svc.List(userID, &currentMonthStart, &now, txType) // this mounth
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total := 0.0
	for _, tx := range list {
		total += tx.Amount
	}

	daysPassed := now.Sub(currentMonthStart).Hours() / 24
	if daysPassed == 0 {
		daysPassed = 1
	}

	average := total / daysPassed
	daysNext := float64(nextMonthEnd.Day())
	predicted := average * daysNext

	c.JSON(http.StatusOK, gin.H{
		"average_per_day": average,
		"days_next_month": daysNext,
		"predicted_total": predicted,
	})
}
