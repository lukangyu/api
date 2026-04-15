package router

import (
	"net/http"
	"os"
	"strings"

	"api_zhuanfa/internal/config"
	"api_zhuanfa/internal/handler"
	"api_zhuanfa/internal/middleware"
	"api_zhuanfa/internal/proxy"
	"api_zhuanfa/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type Services struct {
	AuthSvc     *service.AuthService
	UserSvc     *service.UserService
	ApiKeySvc   *service.ApiKeyService
	UpstreamSvc *service.UpstreamService
	StatsSvc    *service.StatsService
	Logger      *service.RequestLogger
	ProxyEngine *proxy.Engine
}

func New(cfg config.Config, db *gorm.DB, svc Services) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg, svc.UpstreamSvc))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	authHandler := handler.NewAuthHandler(svc.AuthSvc)
	userHandler := handler.NewUserHandler(svc.UserSvc)
	apiKeyHandler := handler.NewAPIKeyHandler(svc.ApiKeySvc)
	upstreamHandler := handler.NewUpstreamHandler(svc.UpstreamSvc)
	logHandler := handler.NewLogHandler(db)
	dashboardHandler := handler.NewDashboardHandler(svc.StatsSvc)

	r.POST("/api/auth/login", authHandler.Login)

	jwtMid := middleware.NewJWTAuth(cfg)
	admin := r.Group("/api/admin")
	admin.Use(jwtMid.Middleware())
	{
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PUT("/users/:id", userHandler.Update)
		admin.DELETE("/users/:id", userHandler.Delete)

		admin.GET("/api-keys", apiKeyHandler.List)
		admin.POST("/api-keys", apiKeyHandler.Create)
		admin.PUT("/api-keys/:id", apiKeyHandler.Update)
		admin.DELETE("/api-keys/:id", apiKeyHandler.Delete)

		admin.GET("/upstreams", upstreamHandler.List)
		admin.POST("/upstreams", upstreamHandler.Create)
		admin.POST("/upstreams/test", upstreamHandler.Test)
		admin.PUT("/upstreams/:id", upstreamHandler.Update)
		admin.DELETE("/upstreams/:id", upstreamHandler.Delete)

		admin.GET("/logs", logHandler.List)
		admin.GET("/stats/overview", dashboardHandler.Overview)
		admin.GET("/stats/daily", dashboardHandler.Daily)
	}

	apiKeyAuth := middleware.NewApiKeyAuth(svc.ApiKeySvc, svc.UpstreamSvc)
	rateLimiter := middleware.NewRateLimiter(rate.Limit(cfg.RateLimitRate), cfg.RateLimitBurst)
	proxyGroup := r.Group("/proxy/:api_name")
	proxyGroup.Use(apiKeyAuth.Middleware(), rateLimiter.Middleware())
	{
		proxyGroup.Any("/*path", proxy.NewHandler(svc.ProxyEngine))
	}

	r.Static("/assets", "./static/assets")
	r.GET("/", serveIndex)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/proxy/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		serveIndex(c)
	})

	return r
}

func serveIndex(c *gin.Context) {
	body, err := os.ReadFile("./static/index.html")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "frontend not built yet"})
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", body)
}
