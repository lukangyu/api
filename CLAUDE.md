# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Backend
make build                    # go build -o bin/api-gateway ./cmd/server
make run                      # go run ./cmd/server (port 8080)
make fmt                      # gofmt
make tidy                     # go mod tidy
go build ./cmd/server/...     # verify compilation (root package has no main)
go vet ./internal/... ./cmd/... # static analysis

# Frontend (web/)
cd web && npm install && npm run build  # outputs to ../static/
cd web && npm run dev                   # dev server on :5173, proxies /api and /proxy to :8080

# Docker
docker compose up -d --build  # multi-stage: node build → go build → alpine runtime
```

## Architecture

Internal API gateway that proxies HTTP requests to configured upstream APIs. Employees get internal API Keys; the gateway injects real upstream credentials on their behalf.

### Request Flow

```
Client → /proxy/:api_name/*path
  → ApiKeyAuth middleware (validates Bearer sk-xxx, sets context)
  → RateLimiter middleware (per-key token bucket)
  → proxy.NewHandler
    → engine.ResolveUpstream(api_name)     # name-based lookup with TTL cache
    → CheckUpstreamAllowed(key, upstream)  # comma-separated ID whitelist
    → engine.BuildProxy                    # httputil.ReverseProxy with custom Director
      → Director: URL rewrite + strip prefix + del Authorization + inject upstream auth + extra headers
    → context.WithTimeout(upstream.TimeoutSeconds)
    → ServeHTTP
    → BuildLogEntry → async batch logger
```

### Layer Responsibilities

- **model/** — GORM structs: User, ApiKey, Upstream, RequestLog. SQLite with single-writer (MaxOpenConns=1).
- **service/** — Business logic + validation. UpstreamService has RWMutex TTL cache. RequestLogger uses buffered channel + goroutine for async batch inserts.
- **handler/** — Thin HTTP handlers. Upstream.Test endpoint builds a real HTTP probe using the same auth/proxy logic as the proxy chain.
- **proxy/** — Reverse proxy core. Director rewrites URL/auth. Transport supports per-upstream HTTP/SOCKS5 proxy via `proxy_url` field. `ApplyUpstreamAuth` and `ApplyExtraHeaders` are exported for reuse by the test handler.
- **middleware/** — Dual auth: JWT for admin panel, API Key for proxy routes. Rate limiter keyed by API Key ID.
- **router/** — Route registration. Admin group under `/api/admin/` (JWT), proxy group under `/proxy/:api_name/*path` (API Key + rate limit). Static `/upstreams/test` must be registered before `/:id` params.

### Key Design Decisions

- Authorization header is **always stripped** before forwarding to upstream (director.go). For `auth_type=bearer/header/query`, the gateway injects upstream credentials. For `auth_type=none`, nothing is injected — `none` means the upstream genuinely needs no auth, not "passthrough caller auth".
- `FlushInterval: -1` on ReverseProxy enables SSE streaming.
- Upstream `proxy_url` creates a per-request Transport (not cached) to support per-upstream HTTP/SOCKS5 proxies.
- Frontend is a React SPA served from `./static/` by the Go server. In dev mode, Vite proxies API calls to Go backend.

## Configuration

All config via environment variables (see `.env.example`). Critical: `JWT_SECRET` and `DEFAULT_ADMIN_PASS` must be changed from defaults.
