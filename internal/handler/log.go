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
	items, err := h.enrichLogRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

type logResponse struct {
	model.RequestLog
	UserName     string `json:"user_name"`
	APIKeyName   string `json:"api_key_name"`
	UpstreamName string `json:"upstream_name"`
}

func (h *LogHandler) enrichLogRows(rows []model.RequestLog) ([]logResponse, error) {
	userIDs := make([]uint, 0, len(rows))
	apiKeyIDs := make([]uint, 0, len(rows))
	upstreamIDs := make([]uint, 0, len(rows))
	seenUsers := make(map[uint]struct{}, len(rows))
	seenKeys := make(map[uint]struct{}, len(rows))
	seenUpstreams := make(map[uint]struct{}, len(rows))

	for _, row := range rows {
		if _, ok := seenUsers[row.UserID]; !ok && row.UserID != 0 {
			seenUsers[row.UserID] = struct{}{}
			userIDs = append(userIDs, row.UserID)
		}
		if _, ok := seenKeys[row.ApiKeyID]; !ok && row.ApiKeyID != 0 {
			seenKeys[row.ApiKeyID] = struct{}{}
			apiKeyIDs = append(apiKeyIDs, row.ApiKeyID)
		}
		if _, ok := seenUpstreams[row.UpstreamID]; !ok && row.UpstreamID != 0 {
			seenUpstreams[row.UpstreamID] = struct{}{}
			upstreamIDs = append(upstreamIDs, row.UpstreamID)
		}
	}

	userNames := make(map[uint]string, len(userIDs))
	if len(userIDs) > 0 {
		var users []model.User
		if err := h.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			if user.DisplayName != "" {
				userNames[user.ID] = user.DisplayName
				continue
			}
			userNames[user.ID] = user.Username
		}
	}

	apiKeyNames := make(map[uint]string, len(apiKeyIDs))
	if len(apiKeyIDs) > 0 {
		var apiKeys []model.ApiKey
		if err := h.db.Where("id IN ?", apiKeyIDs).Find(&apiKeys).Error; err != nil {
			return nil, err
		}
		for _, apiKey := range apiKeys {
			if apiKey.Name != "" {
				apiKeyNames[apiKey.ID] = apiKey.Name
				continue
			}
			apiKeyNames[apiKey.ID] = apiKey.KeyPrefix
		}
	}

	upstreamNames := make(map[uint]string, len(upstreamIDs))
	if len(upstreamIDs) > 0 {
		var upstreams []model.Upstream
		if err := h.db.Where("id IN ?", upstreamIDs).Find(&upstreams).Error; err != nil {
			return nil, err
		}
		for _, upstream := range upstreams {
			if upstream.DisplayName != "" {
				upstreamNames[upstream.ID] = upstream.DisplayName
				continue
			}
			upstreamNames[upstream.ID] = upstream.Name
		}
	}

	items := make([]logResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, logResponse{
			RequestLog:   row,
			UserName:     userNames[row.UserID],
			APIKeyName:   apiKeyNames[row.ApiKeyID],
			UpstreamName: upstreamNames[row.UpstreamID],
		})
	}
	return items, nil
}
