package middleware

import (
	"strings"

	"api_zhuanfa/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.Config) gin.HandlerFunc {
	allowAll := strings.TrimSpace(cfg.CORSOrigins) == "*"
	if allowAll {
		return cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:    []string{"Authorization", "Content-Type"},
			ExposeHeaders:   []string{"Content-Length"},
		})
	}
	origins := strings.Split(cfg.CORSOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}
