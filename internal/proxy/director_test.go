package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api_zhuanfa/internal/model"
)

func TestBuildDirectorOverwritesClientQueryCredential(t *testing.T) {
	upstream := &model.Upstream{
		Name:        "google",
		BaseURL:     "https://www.googleapis.com",
		AuthType:    "query",
		AuthKey:     "key",
		AuthValue:   "upstream-secret",
		StripPrefix: true,
	}

	director, err := BuildDirector(upstream, nil)
	if err != nil {
		t.Fatalf("build director: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://gateway/proxy/google/youtube/v3/search?key=sk-client&part=snippet", nil)
	director(req)

	if req.URL.Scheme != "https" || req.URL.Host != "www.googleapis.com" {
		t.Fatalf("unexpected upstream target: %s://%s", req.URL.Scheme, req.URL.Host)
	}
	if req.URL.Path != "/youtube/v3/search" {
		t.Fatalf("unexpected upstream path: %s", req.URL.Path)
	}
	if got := req.URL.Query().Get("key"); got != "upstream-secret" {
		t.Fatalf("expected upstream key to overwrite client key, got %q", got)
	}
	if got := req.URL.Query().Get("part"); got != "snippet" {
		t.Fatalf("expected business query param to be preserved, got %q", got)
	}
}

func TestBuildDirectorOverwritesClientHeaderCredential(t *testing.T) {
	upstream := &model.Upstream{
		Name:        "custom",
		BaseURL:     "https://example.com/api",
		AuthType:    "header",
		AuthKey:     "X-API-Key",
		AuthValue:   "upstream-secret",
		StripPrefix: true,
	}

	director, err := BuildDirector(upstream, nil)
	if err != nil {
		t.Fatalf("build director: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://gateway/proxy/custom/items", nil)
	req.Header.Set("X-API-Key", "sk-client")
	director(req)

	if req.URL.Path != "/api/items" {
		t.Fatalf("unexpected upstream path: %s", req.URL.Path)
	}
	if got := req.Header.Get("X-API-Key"); got != "upstream-secret" {
		t.Fatalf("expected upstream header to overwrite client header, got %q", got)
	}
}
