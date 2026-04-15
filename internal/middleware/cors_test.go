package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api_zhuanfa/internal/config"
	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
	sqlite "github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newCORSTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Upstream{}); err != nil {
		t.Fatalf("migrate upstreams: %v", err)
	}
	return db
}

func newCORSTestRouter(t *testing.T, upstream *model.Upstream) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := newCORSTestDB(t)
	upstreamSvc := service.NewUpstreamService(db, 30)
	if err := upstreamSvc.Create(upstream); err != nil {
		t.Fatalf("create upstream: %v", err)
	}

	router := gin.New()
	router.Use(CORS(config.Config{CORSOrigins: "http://app.example.com"}, upstreamSvc))
	router.OPTIONS("/proxy/:api_name/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestCORSAllowsEnabledNativeAuthHeaderPreflight(t *testing.T) {
	router := newCORSTestRouter(t, &model.Upstream{
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

	req := httptest.NewRequest(http.MethodOptions, "/proxy/custom/items", nil)
	req.Header.Set("Origin", "http://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "x-api-key,content-type")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	allowedHeaders := recorder.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowedHeaders, "X-Api-Key") {
		t.Fatalf("expected X-Api-Key in allow headers, got %q", allowedHeaders)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://app.example.com" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
}

func TestCORSDoesNotAllowDisabledNativeAuthHeaderPreflight(t *testing.T) {
	router := newCORSTestRouter(t, &model.Upstream{
		Name:           "custom",
		DisplayName:    "Custom API",
		BaseURL:        "https://example.com",
		AuthType:       "header",
		AuthKey:        "X-API-Key",
		AuthValue:      "upstream-secret",
		TimeoutSeconds: 120,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})

	req := httptest.NewRequest(http.MethodOptions, "/proxy/custom/items", nil)
	req.Header.Set("Origin", "http://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "x-api-key,content-type")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	allowedHeaders := recorder.Header().Get("Access-Control-Allow-Headers")
	if strings.Contains(allowedHeaders, "X-Api-Key") {
		t.Fatalf("expected X-Api-Key to stay blocked, got %q", allowedHeaders)
	}
}
