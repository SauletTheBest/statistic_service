package handler

import (
	"net/http"
	"time"

	"statistic_service/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct{ svc service.TransactionService }

func NewStatsHandler(s service.TransactionService) *StatsHandler { return &StatsHandler{s} }

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
	inc, exp, _ := h.svc.Summary(userID, from, to)
	c.JSON(http.StatusOK, gin.H{"income": inc, "expense": exp})
}
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
	data, _ := h.svc.ByCategory(userID, from, to)
	c.JSON(http.StatusOK, data)
}
