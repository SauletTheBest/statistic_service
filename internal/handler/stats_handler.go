package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/service"
	"statistic_service/pkg/utils" // <-- ДОБАВЛЕНО: Импорт для SendErrorResponse

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type StatsHandler struct {
	svc    service.TransactionService
	logger *logrus.Logger
}

func NewStatsHandler(s service.TransactionService, logger *logrus.Logger) *StatsHandler {
	return &StatsHandler{svc: s, logger: logger}
}

// Summary godoc
// @Summary Get transactions summary
// @Description Returns total income and expenses for the authenticated user
// @Tags Statistics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)"
// @Param date_to query string false "End date in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} map[string]string "error: invalid date format"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /stats/summary [get]
func (h *StatsHandler) Summary(c *gin.Context) {
	userID := c.GetString("userID")
	var from, to *time.Time
	if f := c.Query("date_from"); f != "" {
		t, err := time.Parse(time.RFC3339, f)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_from format in summary request")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_from' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		from = &t
	}
	if tstr := c.Query("date_to"); tstr != "" {
		t, err := time.Parse(time.RFC3339, tstr)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_to format in summary request")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_to' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		to = &t
	}
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to}).Info("Fetching summary stats")
	inc, exp, err := h.svc.Summary(userID, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch summary stats")
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.logger.WithFields(logrus.Fields{"income": inc, "expense": exp}).Info("Summary stats retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"income": inc, "expense": exp})
}

// ByCategory godoc
// @Summary Get summary by category
// @Description Returns sum of transactions grouped by category for the authenticated user
// @Tags Statistics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)"
// @Param date_to query string false "End date in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} map[string]string "error: invalid date format"
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Failure 500 {object} map[string]string "error: internal server error"
// @Security BearerAuth
// @Router /stats/categories [get]
func (h *StatsHandler) ByCategory(c *gin.Context) {
	userID := c.GetString("userID")
	var from, to *time.Time
	if f := c.Query("date_from"); f != "" {
		t, err := time.Parse(time.RFC3339, f)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_from format in ByCategory request")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_from' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		from = &t
	}
	if tstr := c.Query("date_to"); tstr != "" {
		t, err := time.Parse(time.RFC3339, tstr)
		if err != nil {
			h.logger.WithError(err).Warn("Invalid date_to format in ByCategory request")
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid 'date_to' format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
			return
		}
		to = &t
	}
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to}).Info("Fetching stats by category")
	data, err := h.svc.ByCategory(userID, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch stats by category")
		utils.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.logger.WithField("categoriesCount", len(data)).Info("Stats by category retrieved successfully")
	c.JSON(http.StatusOK, data)
}
