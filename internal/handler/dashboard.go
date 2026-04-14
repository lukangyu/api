package handler

import (
	"net/http"
	"strconv"

	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	statsSvc *service.StatsService
}

func NewDashboardHandler(statsSvc *service.StatsService) *DashboardHandler {
	return &DashboardHandler{statsSvc: statsSvc}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	out, err := h.statsSvc.Overview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *DashboardHandler) Daily(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	rows, err := h.statsSvc.Daily(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
