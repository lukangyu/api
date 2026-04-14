package middleware

import (
	"net/http"
	"strings"
	"time"

	"api_zhuanfa/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxAdminUserID = "admin_user_id"
	CtxAdminRole   = "admin_role"
)

type JWTAuth struct {
	cfg config.Config
}

func NewJWTAuth(cfg config.Config) *JWTAuth {
	return &JWTAuth{cfg: cfg}
}

func (j *JWTAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimSpace(authHeader[len("Bearer "):])
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(j.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}
		c.Set(CtxAdminUserID, claims.Subject)
		c.Set(CtxAdminRole, "admin")
		c.Next()
	}
}
