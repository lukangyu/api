package service

import (
	"testing"

	"api_zhuanfa/internal/model"
)

func TestUpstreamServicePrepareTrimsAuthValue(t *testing.T) {
	svc := NewUpstreamService(nil, 30)
	upstream := &model.Upstream{
		Name:           "producthunt",
		DisplayName:    "Product Hunt",
		BaseURL:        " https://api.producthunt.com ",
		AuthType:       " bearer ",
		AuthValue:      " ph_token_with_space \n",
		TimeoutSeconds: 120,
		ExtraHeaders:   "{}",
	}

	if err := svc.Prepare(upstream); err != nil {
		t.Fatalf("prepare upstream: %v", err)
	}

	if upstream.BaseURL != "https://api.producthunt.com" {
		t.Fatalf("expected trimmed base_url, got %q", upstream.BaseURL)
	}
	if upstream.AuthType != "bearer" {
		t.Fatalf("expected trimmed auth_type, got %q", upstream.AuthType)
	}
	if upstream.AuthValue != "ph_token_with_space" {
		t.Fatalf("expected trimmed auth_value, got %q", upstream.AuthValue)
	}
}
