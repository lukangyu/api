package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort       string
	DBPath           string
	JWTSecret        string
	JWTExpireHours   int
	DefaultAdminUser string
	DefaultAdminPass string
	CORSOrigins      string
	UpstreamCacheTTL int
	LoggerBufferSize int
	LoggerFlushSize  int

	RateLimitRate  float64
	RateLimitBurst int
}

func Load() Config {
	return Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DBPath:           getEnv("DB_PATH", "./data/gateway.db"),
		JWTSecret:        getEnv("JWT_SECRET", "change_me_jwt_secret"),
		JWTExpireHours:   getEnvInt("JWT_EXPIRE_HOURS", 24),
		DefaultAdminUser: getEnv("DEFAULT_ADMIN_USER", "admin"),
		DefaultAdminPass: getEnv("DEFAULT_ADMIN_PASS", "changeme123"),
		CORSOrigins:      getEnv("CORS_ORIGINS", "*"),
		UpstreamCacheTTL: getEnvInt("UPSTREAM_CACHE_TTL_SECONDS", 30),
		LoggerBufferSize: getEnvInt("LOGGER_BUFFER_SIZE", 1000),
		LoggerFlushSize:  getEnvInt("LOGGER_FLUSH_SIZE", 100),

		RateLimitRate:  getEnvFloat("RATE_LIMIT_RATE", 20),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 40),
	}
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

func getEnvFloat(key string, defaultValue float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultValue
	}
	return f
}
