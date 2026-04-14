package handler

import (
	"net/http"
	"strconv"
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	db *gorm.DB
}

func NewLogHandler(db *gorm.DB) *LogHandler {
	return &LogHandler{db: db}
}

func (h *LogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	q := h.db.Model(&model.RequestLog{})
	if userID := c.Query("user_id"); userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if upstreamID := c.Query("upstream_id"); upstreamID != "" {
		q = q.Where("upstream_id = ?", upstreamID)
	}
	if status := c.Query("status_code"); status != "" {
		q = q.Where("status_code = ?", status)
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("created_at <= ?", t)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var rows []model.RequestLog
	if err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows, "total": total})
}
