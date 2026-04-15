package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"api_zhuanfa/internal/model"
	proxyutil "api_zhuanfa/internal/proxy"
	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

type UpstreamHandler struct {
	upstreamSvc *service.UpstreamService
}

func NewUpstreamHandler(upstreamSvc *service.UpstreamService) *UpstreamHandler {
	return &UpstreamHandler{upstreamSvc: upstreamSvc}
}

func (h *UpstreamHandler) List(c *gin.Context) {
	rows, err := h.upstreamSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *UpstreamHandler) Create(c *gin.Context) {
	var req model.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.upstreamSvc.Create(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": req})
}

func (h *UpstreamHandler) Test(c *gin.Context) {
	var req model.Upstream
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.upstreamSvc.Prepare(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var transport *http.Transport
	var err error
	if req.ProxyURL == "" {
		transport = proxyutil.NewTransport()
	} else {
		transport, err = proxyutil.NewTransportWithProxy(req.ProxyURL)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	httpReq, targetURL, err := buildUpstreamTestRequest(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := &http.Client{
		Timeout:   time.Duration(req.TimeoutSeconds) * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"ok":         false,
			"reachable":  false,
			"category":   "network_error",
			"message":    err.Error(),
			"target_url": targetURL,
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	excerpt := strings.TrimSpace(string(body))
	category, ok, message := classifyUpstreamTestResponse(resp.StatusCode)
	c.JSON(http.StatusOK, gin.H{
		"ok":               ok,
		"reachable":        true,
		"category":         category,
		"message":          message,
		"target_url":       targetURL,
		"status_code":      resp.StatusCode,
		"response_excerpt": excerpt,
	})
}

func buildUpstreamTestRequest(up *model.Upstream) (*http.Request, string, error) {
	parsed, err := url.Parse(up.BaseURL)
	if err != nil {
		return nil, "", err
	}

	method := http.MethodGet
	targetURL := parsed.String()
	var body io.Reader

	if isProductHuntUpstream(up, parsed) {
		method = http.MethodPost
		targetURL = buildProductHuntTestURL(parsed)
		body = bytes.NewBufferString(`{"query":"query { posts(first: 1) { edges { node { id } } } }"}`)
	} else if isDoubaoEmbeddingUpstream(up, parsed) {
		method = http.MethodPost
		targetURL = buildDoubaoEmbeddingTestURL(parsed)
		body = bytes.NewBufferString(`{"model":"doubao-embedding-vision-251215","input":[{"type":"text","text":"ping"}],"dimensions":2048}`)
	} else if strings.Contains(strings.ToLower(parsed.Path), "graphql") {
		method = http.MethodPost
		body = bytes.NewBufferString(`{"query":"query { __typename }"}`)
	}

	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, "", err
	}

	if method == http.MethodPost {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
	}

	applyUpstreamTestAuth(req, up)
	applyUpstreamTestExtraHeaders(req, up.ExtraHeaders)
	return req, targetURL, nil
}

func isProductHuntUpstream(up *model.Upstream, parsed *url.URL) bool {
	fingerprint := strings.ToLower(strings.Join([]string{
		up.Name,
		up.DisplayName,
		up.BaseURL,
		parsed.Host,
	}, " "))
	return strings.Contains(fingerprint, "producthunt") || strings.Contains(fingerprint, "product hunt")
}

func buildProductHuntTestURL(parsed *url.URL) string {
	if strings.TrimSpace(parsed.Path) == "" || parsed.Path == "/" {
		cloned := *parsed
		cloned.Path = "/v2/api/graphql"
		return cloned.String()
	}
	return parsed.String()
}

func isDoubaoEmbeddingUpstream(up *model.Upstream, parsed *url.URL) bool {
	fingerprint := strings.ToLower(strings.Join([]string{
		up.Name,
		up.DisplayName,
		up.BaseURL,
		parsed.Host,
		parsed.Path,
	}, " "))
	return strings.Contains(fingerprint, "embeddings/multimodal") ||
		strings.Contains(fingerprint, "doubao_embedding") ||
		(strings.Contains(fingerprint, "doubao") && strings.Contains(fingerprint, "embedding"))
}

func buildDoubaoEmbeddingTestURL(parsed *url.URL) string {
	if strings.TrimSpace(parsed.Path) == "" || parsed.Path == "/" {
		cloned := *parsed
		cloned.Path = "/api/v3/embeddings/multimodal"
		return cloned.String()
	}
	return parsed.String()
}

func (h *UpstreamHandler) Update(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.upstreamSvc.Update(uint(id64), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *UpstreamHandler) Delete(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.upstreamSvc.Delete(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func applyUpstreamTestAuth(req *http.Request, up *model.Upstream) {
	switch strings.ToLower(strings.TrimSpace(up.AuthType)) {
	case "bearer":
		if strings.TrimSpace(up.AuthValue) != "" {
			req.Header.Set("Authorization", "Bearer "+up.AuthValue)
		}
	case "header":
		if strings.TrimSpace(up.AuthKey) != "" {
			req.Header.Set(up.AuthKey, up.AuthValue)
		}
	case "query":
		if strings.TrimSpace(up.AuthKey) != "" {
			q := req.URL.Query()
			q.Set(up.AuthKey, up.AuthValue)
			req.URL.RawQuery = q.Encode()
		}
	}
}

func applyUpstreamTestExtraHeaders(req *http.Request, raw string) {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "{}" {
		return
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		return
	}
	for k, v := range headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
}

func classifyUpstreamTestResponse(statusCode int) (category string, ok bool, message string) {
	switch {
	case statusCode >= 200 && statusCode < 400:
		return "success", true, "连通成功"
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return "auth_error", true, "网络可达，但鉴权可能有误（请检查 auth_value）"
	case statusCode == http.StatusNotFound:
		return "path_error", true, "网络可达（根路径返回 404 属正常，实际转发路径由调用方 URL 决定）"
	case statusCode == http.StatusMethodNotAllowed:
		return "method_error", true, "网络可达（根路径不支持 GET 属正常）"
	default:
		return "http_error", true, "网络可达，上游返回了非常规状态码"
	}
}
