package handler

import (
	"net/http"
	"strconv"
	"time"

	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeySvc *service.ApiKeyService
}

func NewAPIKeyHandler(apiKeySvc *service.ApiKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeySvc: apiKeySvc}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req struct {
		UserID           uint    `json:"user_id"`
		Name             string  `json:"name"`
		RequestLimit     int64   `json:"request_limit"`
		ExpiresAt        *string `json:"expires_at"`
		AllowedUpstreams string  `json:"allowed_upstreams"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be RFC3339"})
			return
		}
		expiresAt = &t
	}
	plain, item, err := h.apiKeySvc.Generate(req.UserID, req.Name, req.RequestLimit, expiresAt, req.AllowedUpstreams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plain_key": plain, "item": item})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	rows, total, err := h.apiKeySvc.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows, "total": total})
}

func (h *APIKeyHandler) Update(c *gin.Context) {
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
	if err := h.apiKeySvc.Update(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.apiKeySvc.Revoke(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
