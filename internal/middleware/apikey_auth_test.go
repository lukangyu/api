package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newApiKeyAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ApiKey{}, &model.Upstream{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}
	return db
}

func newApiKeyAuthTestHarness(t *testing.T, upstream *model.Upstream) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := newApiKeyAuthTestDB(t)
	user := &model.User{
		Username:     "tester",
		PasswordHash: "x",
		DisplayName:  "Tester",
		Role:         "user",
		IsActive:     true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	apiKeySvc := service.NewApiKeyService(db)
	plainKey, _, err := apiKeySvc.Generate(user.ID, "demo", 0, nil, "")
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	upstreamSvc := service.NewUpstreamService(db, 30)
	if err := upstreamSvc.Create(upstream); err != nil {
		t.Fatalf("create upstream: %v", err)
	}

	auth := NewApiKeyAuth(apiKeySvc, upstreamSvc)
	router := gin.New()
	router.GET("/proxy/:api_name/*path", auth.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return router, plainKey
}

func decodeErrorMessage(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error
}

func TestApiKeyAuthAcceptsBearerToken(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:           "google",
		DisplayName:    "Google API",
		BaseURL:        "https://www.googleapis.com",
		AuthType:       "query",
		AuthKey:        "key",
		AuthValue:      "upstream-secret",
		TimeoutSeconds: 120,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/google/youtube/v3/search", nil)
	req.Header.Set("Authorization", "Bearer "+plainKey)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestApiKeyAuthAcceptsNativeQueryKeyWhenEnabled(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:                  "google",
		DisplayName:           "Google API",
		BaseURL:               "https://www.googleapis.com",
		AuthType:              "query",
		AuthKey:               "key",
		AuthValue:             "upstream-secret",
		AllowNativeClientAuth: true,
		TimeoutSeconds:        120,
		ExtraHeaders:          "{}",
		IsActive:              true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/google/youtube/v3/search?key="+plainKey, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestApiKeyAuthRejectsNativeQueryKeyWhenDisabled(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:           "google",
		DisplayName:    "Google API",
		BaseURL:        "https://www.googleapis.com",
		AuthType:       "query",
		AuthKey:        "key",
		AuthValue:      "upstream-secret",
		TimeoutSeconds: 120,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/google/youtube/v3/search?key="+plainKey, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := decodeErrorMessage(t, recorder); got != "missing bearer token" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestApiKeyAuthAcceptsNativeHeaderWhenEnabled(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:                  "custom",
		DisplayName:           "Custom API",
		BaseURL:               "https://example.com",
		AuthType:              "header",
		AuthKey:               "X-API-Key",
		AuthValue:             "upstream-secret",
		AllowNativeClientAuth: true,
		TimeoutSeconds:        120,
		ExtraHeaders:          "{}",
		IsActive:              true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/custom/resource", nil)
	req.Header.Set("X-API-Key", plainKey)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestApiKeyAuthAcceptsRawAuthorizationHeaderWhenHeaderCompatEnabled(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:                  "legacy-auth",
		DisplayName:           "Legacy Auth API",
		BaseURL:               "https://example.com",
		AuthType:              "header",
		AuthKey:               "Authorization",
		AuthValue:             "upstream-secret",
		AllowNativeClientAuth: true,
		TimeoutSeconds:        120,
		ExtraHeaders:          "{}",
		IsActive:              true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/legacy-auth/resource", nil)
	req.Header.Set("Authorization", plainKey)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestApiKeyAuthPrefersBearerTokenOverNativeCompatibility(t *testing.T) {
	router, plainKey := newApiKeyAuthTestHarness(t, &model.Upstream{
		Name:                  "google",
		DisplayName:           "Google API",
		BaseURL:               "https://www.googleapis.com",
		AuthType:              "query",
		AuthKey:               "key",
		AuthValue:             "upstream-secret",
		AllowNativeClientAuth: true,
		TimeoutSeconds:        120,
		ExtraHeaders:          "{}",
		IsActive:              true,
	})

	req := httptest.NewRequest(http.MethodGet, "/proxy/google/youtube/v3/search?key="+plainKey, nil)
	req.Header.Set("Authorization", "Bearer sk-invalid")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := decodeErrorMessage(t, recorder); got != "invalid api key" {
		t.Fatalf("unexpected error message: %q", got)
	}
}
