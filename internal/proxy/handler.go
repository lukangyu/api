package proxy

import (
	"net/http"
	"time"

	"api_zhuanfa/internal/middleware"
	"api_zhuanfa/internal/model"
	"github.com/gin-gonic/gin"
)

func NewHandler(engine *Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiName := c.Param("api_name")
		upstream, err := engine.ResolveUpstream(apiName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "upstream not found"})
			return
		}

		keyValue, ok := c.Get(middleware.CtxApiKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "api key missing"})
			return
		}
		apiKey, ok := keyValue.(*model.ApiKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "api key invalid"})
			return
		}
		if !engine.apiKeySvc.CheckUpstreamAllowed(apiKey, upstream.ID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "api key is not allowed for this upstream"})
			return
		}

		meta := &MetaCarrier{StartedAt: time.Now()}
		p, err := engine.BuildProxy(upstream, meta)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		apiKeyID, _ := c.Get(middleware.CtxApiKeyID)
		userID, _ := c.Get(middleware.CtxUserID)

		p.ServeHTTP(c.Writer, c.Request)

		apiKeyUint := toUint(apiKeyID)
		userUint := toUint(userID)
		log := BuildLogEntry(meta, c.Request, apiKeyUint, userUint, upstream.ID)
		engine.AfterProxy(log)
	}
}

func toUint(v interface{}) uint {
	switch t := v.(type) {
	case uint:
		return t
	case int:
		if t < 0 {
			return 0
		}
		return uint(t)
	default:
		return 0
	}
}
