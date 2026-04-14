package middleware

import (
	"net/http"
	"strings"

	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	CtxApiKeyID = "api_key_id"
	CtxUserID   = "user_id"
	CtxApiKey   = "api_key"
	CtxUser     = "user"
	CtxApiName  = "api_name"
)

type ApiKeyAuth struct {
	apiKeyService *service.ApiKeyService
}

func NewApiKeyAuth(apiKeyService *service.ApiKeyService) *ApiKeyAuth {
	return &ApiKeyAuth{apiKeyService: apiKeyService}
}

func (a *ApiKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		plainKey := strings.TrimSpace(authHeader[len("Bearer "):])
		if plainKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "empty api key"})
			return
		}
		key, user, err := a.apiKeyService.Validate(plainKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(CtxApiKeyID, key.ID)
		c.Set(CtxUserID, user.ID)
		c.Set(CtxApiKey, key)
		c.Set(CtxUser, user)
		c.Next()
	}
}
