package main

import (
	"log"

	"api_zhuanfa/internal/config"
	"api_zhuanfa/internal/database"
	"api_zhuanfa/internal/proxy"
	"api_zhuanfa/internal/router"
	"api_zhuanfa/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}

	if err := database.SeedAdmin(db, cfg.DefaultAdminUser, cfg.DefaultAdminPass); err != nil {
		log.Fatalf("seed admin failed: %v", err)
	}

	authSvc := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpireHours)
	userSvc := service.NewUserService(db)
	apiKeySvc := service.NewApiKeyService(db)
	upstreamSvc := service.NewUpstreamService(db, cfg.UpstreamCacheTTL)
	statsSvc := service.NewStatsService(db)
	logger := service.NewRequestLogger(db, cfg.LoggerBufferSize, cfg.LoggerFlushSize)
	defer logger.Close()

	engine := proxy.NewEngine(upstreamSvc, apiKeySvc, logger)

	r := router.New(cfg, db, router.Services{
		AuthSvc:     authSvc,
		UserSvc:     userSvc,
		ApiKeySvc:   apiKeySvc,
		UpstreamSvc: upstreamSvc,
		StatsSvc:    statsSvc,
		Logger:      logger,
		ProxyEngine: engine,
	})

	log.Printf("server started on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
