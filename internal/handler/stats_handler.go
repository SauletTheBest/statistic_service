package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/service"

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
// @Param date_from query string false "Start date in RFC3339 format"
// @Param date_to query string false "End date in RFC3339 format"
// @Success 200 {object} map[string]float64
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Security BearerAuth
// @Router /stats/summary [get]
func (h *StatsHandler) Summary(c *gin.Context) {
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
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to}).Info("Fetching summary stats")
	inc, exp, err := h.svc.Summary(userID, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch summary stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Param date_from query string false "Start date in RFC3339 format"
// @Param date_to query string false "End date in RFC3339 format"
// @Success 200 {object} map[string]float64
// @Failure 401 {object} map[string]string "error: unauthorized"
// @Security BearerAuth
// @Router /stats/categories [get]
func (h *StatsHandler) ByCategory(c *gin.Context) {
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
	h.logger.WithFields(logrus.Fields{"userID": userID, "from": from, "to": to}).Info("Fetching stats by category")
	data, err := h.svc.ByCategory(userID, from, to)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch stats by category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logger.WithField("categoriesCount", len(data)).Info("Stats by category retrieved successfully")
	c.JSON(http.StatusOK, data)
}
