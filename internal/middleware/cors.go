package middleware

import (
	"net/http"
	"strings"

	"api_zhuanfa/internal/config"
	"api_zhuanfa/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var defaultCORSMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
var defaultCORSHeaders = []string{"Authorization", "Content-Type"}

func CORS(cfg config.Config, upstreamSvc *service.UpstreamService) gin.HandlerFunc {
	base := cors.New(buildCORSConfig(cfg))
	allowAll := strings.TrimSpace(cfg.CORSOrigins) == "*"
	allowedOrigins := splitCORSOrigins(cfg.CORSOrigins)

	return func(c *gin.Context) {
		if handleNativeAuthPreflight(c, upstreamSvc, allowAll, allowedOrigins) {
			return
		}
		base(c)
	}
}

func buildCORSConfig(cfg config.Config) cors.Config {
	allowAll := strings.TrimSpace(cfg.CORSOrigins) == "*"
	if allowAll {
		return cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    defaultCORSMethods,
			AllowHeaders:    defaultCORSHeaders,
			ExposeHeaders:   []string{"Content-Length"},
		}
	}
	return cors.Config{
		AllowOrigins:     splitCORSOrigins(cfg.CORSOrigins),
		AllowMethods:     defaultCORSMethods,
		AllowHeaders:     defaultCORSHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
}

func splitCORSOrigins(raw string) []string {
	origins := strings.Split(raw, ",")
	filtered := make([]string, 0, len(origins))
	for i := range origins {
		origin := strings.TrimSpace(origins[i])
		if origin == "" {
			continue
		}
		filtered = append(filtered, origin)
	}
	return filtered
}

func handleNativeAuthPreflight(c *gin.Context, upstreamSvc *service.UpstreamService, allowAll bool, allowedOrigins []string) bool {
	if c.Request.Method != http.MethodOptions {
		return false
	}
	if upstreamSvc == nil {
		return false
	}

	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		return false
	}
	if !allowAll && !originAllowed(origin, allowedOrigins) {
		return false
	}

	requestMethod := strings.TrimSpace(c.GetHeader("Access-Control-Request-Method"))
	requestHeaders := normalizeHeaderList(c.GetHeader("Access-Control-Request-Headers"))
	if requestMethod == "" || len(requestHeaders) == 0 {
		return false
	}

	apiName := proxyAPIName(c.Request.URL.Path)
	if apiName == "" {
		return false
	}

	upstream, err := upstreamSvc.GetActiveByName(apiName)
	if err != nil || upstream == nil {
		return false
	}
	if !upstream.AllowNativeClientAuth || !strings.EqualFold(strings.TrimSpace(upstream.AuthType), "header") {
		return false
	}

	authHeader := strings.TrimSpace(upstream.AuthKey)
	if authHeader == "" || !containsHeader(requestHeaders, authHeader) {
		return false
	}

	headers := canonicalizeHeaderList(append(append([]string{}, defaultCORSHeaders...), authHeader))
	header := c.Writer.Header()
	if allowAll {
		header.Set("Access-Control-Allow-Origin", "*")
	} else {
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Set("Vary", "Origin")
		header.Add("Vary", "Access-Control-Request-Method")
		header.Add("Vary", "Access-Control-Request-Headers")
	}
	header.Set("Access-Control-Allow-Methods", strings.Join(defaultCORSMethods, ","))
	header.Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	header.Set("Access-Control-Expose-Headers", "Content-Length")
	c.AbortWithStatus(http.StatusNoContent)
	return true
}

func proxyAPIName(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[0] != "proxy" {
		return ""
	}
	return parts[1]
}

func originAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}

func normalizeHeaderList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		header := http.CanonicalHeaderKey(strings.TrimSpace(part))
		if header == "" {
			continue
		}
		key := strings.ToLower(header)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, header)
	}
	return out
}

func canonicalizeHeaderList(headers []string) []string {
	return normalizeHeaderList(strings.Join(headers, ","))
}

func containsHeader(headers []string, target string) bool {
	target = http.CanonicalHeaderKey(strings.TrimSpace(target))
	for _, header := range headers {
		if strings.EqualFold(header, target) {
			return true
		}
	}
	return false
}
