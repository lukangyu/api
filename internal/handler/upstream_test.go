package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
	"github.com/gin-gonic/gin"
)

func TestUpstreamHandlerTestInjectsBearerAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer upstream-secret" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstreamServer.Close()

	handler := NewUpstreamHandler(service.NewUpstreamService(nil, 30))
	reqBody, err := json.Marshal(model.Upstream{
		Name:           "doubao_embedding",
		DisplayName:    "Doubao Embedding",
		BaseURL:        upstreamServer.URL,
		AuthType:       "bearer",
		AuthValue:      "upstream-secret",
		TimeoutSeconds: 5,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("marshal upstream request: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/test", bytes.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Test(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		OK         bool   `json:"ok"`
		Reachable  bool   `json:"reachable"`
		Category   string `json:"category"`
		StatusCode int    `json:"status_code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK || !resp.Reachable {
		t.Fatalf("expected successful connectivity result, got %+v", resp)
	}
	if resp.Category != "success" {
		t.Fatalf("unexpected category: %q", resp.Category)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected upstream status code: %d", resp.StatusCode)
	}
}

func TestUpstreamHandlerTestUsesProductHuntGraphQLProbe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST probe, got %s", r.Method)
		}
		if r.URL.Path != "/v2/api/graphql" {
			t.Fatalf("expected Product Hunt GraphQL path, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ph-secret" {
			t.Fatalf("expected bearer auth, got %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected JSON content type, got %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("posts(first: 1)")) {
			t.Fatalf("expected Product Hunt probe query, got %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"posts":{"edges":[]}}}`))
	}))
	defer upstreamServer.Close()

	handler := NewUpstreamHandler(service.NewUpstreamService(nil, 30))
	reqBody, err := json.Marshal(model.Upstream{
		Name:           "producthunt",
		DisplayName:    "Product Hunt GraphQL",
		BaseURL:        upstreamServer.URL,
		AuthType:       "bearer",
		AuthValue:      "ph-secret",
		TimeoutSeconds: 5,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("marshal upstream request: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/test", bytes.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Test(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		OK         bool   `json:"ok"`
		Reachable  bool   `json:"reachable"`
		Category   string `json:"category"`
		StatusCode int    `json:"status_code"`
		TargetURL  string `json:"target_url"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK || !resp.Reachable {
		t.Fatalf("expected successful connectivity result, got %+v", resp)
	}
	if resp.Category != "success" {
		t.Fatalf("unexpected category: %q", resp.Category)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected upstream status code: %d", resp.StatusCode)
	}
	if resp.TargetURL != upstreamServer.URL+"/v2/api/graphql" {
		t.Fatalf("unexpected target url: %s", resp.TargetURL)
	}
}

func TestUpstreamHandlerTestUsesDoubaoEmbeddingProbe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST probe, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/embeddings/multimodal" {
			t.Fatalf("expected Doubao embedding path, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ark-secret" {
			t.Fatalf("expected bearer auth, got %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected JSON content type, got %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("doubao-embedding-vision-251215")) {
			t.Fatalf("expected Doubao probe model, got %s", string(body))
		}
		if !bytes.Contains(body, []byte(`"dimensions":2048`)) {
			t.Fatalf("expected dimensions in probe body, got %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2]}]}`))
	}))
	defer upstreamServer.Close()

	handler := NewUpstreamHandler(service.NewUpstreamService(nil, 30))
	reqBody, err := json.Marshal(model.Upstream{
		Name:           "doubao_embedding",
		DisplayName:    "Doubao Embedding",
		BaseURL:        upstreamServer.URL,
		AuthType:       "bearer",
		AuthValue:      "ark-secret",
		TimeoutSeconds: 5,
		ExtraHeaders:   "{}",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("marshal upstream request: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/test", bytes.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Test(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var resp struct {
		OK         bool   `json:"ok"`
		Reachable  bool   `json:"reachable"`
		Category   string `json:"category"`
		StatusCode int    `json:"status_code"`
		TargetURL  string `json:"target_url"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK || !resp.Reachable {
		t.Fatalf("expected successful connectivity result, got %+v", resp)
	}
	if resp.Category != "success" {
		t.Fatalf("unexpected category: %q", resp.Category)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected upstream status code: %d", resp.StatusCode)
	}
	if resp.TargetURL != upstreamServer.URL+"/api/v3/embeddings/multimodal" {
		t.Fatalf("unexpected target url: %s", resp.TargetURL)
	}
}
