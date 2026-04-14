package handler

import (
	"net/http"
	"strconv"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

type UpstreamHandler struct {
	upstreamSvc *service.UpstreamService
}

func NewUpstreamHandler(upstreamSvc *service.UpstreamService) *UpstreamHandler {
	return &UpstreamHandler{upstreamSvc: upstreamSvc}
}

func (h *UpstreamHandler) List(c *gin.Context) {
	rows, err := h.upstreamSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *UpstreamHandler) Create(c *gin.Context) {
	var req model.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.upstreamSvc.Create(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": req})
}

func (h *UpstreamHandler) Update(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.upstreamSvc.Update(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *UpstreamHandler) Delete(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.upstreamSvc.Delete(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
