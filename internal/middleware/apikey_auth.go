package middleware

import (
	"errors"
	"net/http"
	"strings"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	upstreamSvc   *service.UpstreamService
}

func NewApiKeyAuth(apiKeyService *service.ApiKeyService, upstreamSvc *service.UpstreamService) *ApiKeyAuth {
	return &ApiKeyAuth{apiKeyService: apiKeyService, upstreamSvc: upstreamSvc}
}

func (a *ApiKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiName := strings.TrimSpace(c.Param("api_name"))
		if apiName != "" {
			c.Set(CtxApiName, apiName)
		}

		plainKey, hasBearer := extractBearerToken(c.GetHeader("Authorization"))
		if !hasBearer {
			upstream, err := a.lookupUpstream(apiName)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "load upstream failed"})
				return
			}
			if upstream != nil && upstream.AllowNativeClientAuth {
				var found bool
				plainKey, found = extractNativeClientAPIKey(c, upstream)
				if !found {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
					return
				}
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
				return
			}
		}

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

func extractBearerToken(authHeader string) (string, bool) {
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return "", false
	}
	return strings.TrimSpace(authHeader[len("Bearer "):]), true
}

func extractNativeClientAPIKey(c *gin.Context, upstream *model.Upstream) (string, bool) {
	if upstream == nil {
		return "", false
	}

	switch strings.ToLower(strings.TrimSpace(upstream.AuthType)) {
	case "query":
		if strings.TrimSpace(upstream.AuthKey) == "" {
			return "", false
		}
		value, found := c.GetQuery(upstream.AuthKey)
		return strings.TrimSpace(value), found
	case "header":
		if strings.TrimSpace(upstream.AuthKey) == "" {
			return "", false
		}
		value := c.GetHeader(upstream.AuthKey)
		if value == "" {
			return "", false
		}
		return strings.TrimSpace(value), true
	default:
		return "", false
	}
}

func (a *ApiKeyAuth) lookupUpstream(apiName string) (*model.Upstream, error) {
	if a.upstreamSvc == nil || strings.TrimSpace(apiName) == "" {
		return nil, nil
	}

	upstream, err := a.upstreamSvc.GetActiveByName(apiName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return upstream, nil
}
