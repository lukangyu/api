package handler

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"api_zhuanfa/internal/model"
	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newLogHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "logs.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ApiKey{}, &model.Upstream{}, &model.RequestLog{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestLogHandlerListReturnsReadableNames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newLogHandlerTestDB(t)
	user := model.User{ID: 1, Username: "alice", DisplayName: "Alice", PasswordHash: "hash", Role: "admin", IsActive: true}
	apiKey := model.ApiKey{ID: 1, UserID: 1, KeyHash: "hash", KeyPrefix: "sk-demo", Name: "研发调用", IsActive: true}
	upstream := model.Upstream{ID: 1, Name: "doubao_embedding", DisplayName: "Doubao Embedding", BaseURL: "https://example.com", AuthType: "none", IsActive: true}
	logEntry := model.RequestLog{
		ID:            1,
		ApiKeyID:      1,
		UserID:        1,
		UpstreamID:    1,
		Method:        "POST",
		Path:          "/proxy/doubao_embedding/api/v3/embeddings/multimodal",
		UpstreamPath:  "/api/v3/embeddings/multimodal",
		StatusCode:    200,
		RequestBytes:  128,
		ResponseBytes: 256,
		LatencyMs:     42,
		ClientIP:      "127.0.0.1",
		CreatedAt:     time.Now(),
	}
	for _, entity := range []interface{}{&user, &apiKey, &upstream, &logEntry} {
		if err := db.Create(entity).Error; err != nil {
			t.Fatalf("seed test data: %v", err)
		}
	}

	handler := NewLogHandler(db)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/admin/logs", nil)

	handler.List(ctx)

	if recorder.Code != 200 {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		Items []struct {
			UserName     string `json:"user_name"`
			APIKeyName   string `json:"api_key_name"`
			UpstreamName string `json:"upstream_name"`
		} `json:"items"`
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("unexpected response shape: total=%d items=%d", resp.Total, len(resp.Items))
	}
	if resp.Items[0].UserName != "Alice" {
		t.Fatalf("unexpected user_name: %q", resp.Items[0].UserName)
	}
	if resp.Items[0].APIKeyName != "研发调用" {
		t.Fatalf("unexpected api_key_name: %q", resp.Items[0].APIKeyName)
	}
	if resp.Items[0].UpstreamName != "Doubao Embedding" {
		t.Fatalf("unexpected upstream_name: %q", resp.Items[0].UpstreamName)
	}
}
