# Native Client Auth Compatibility

## Goal

Add an opt-in compatibility mode per upstream so existing clients can keep using the upstream's native auth location while the gateway still validates employee API keys and injects upstream credentials on the outbound request.

## Scope

- Add `allow_native_client_auth` to upstreams, defaulting to `false`
- Keep `Authorization: Bearer sk-...` as the default client auth format
- When compatibility is enabled for an upstream:
  - `query` upstreams may accept the employee API key from the configured query parameter
  - `header` upstreams may accept the employee API key from the configured header
  - `bearer` upstreams continue to use `Authorization: Bearer sk-...`
- Preserve upstream allowlist checks, rate limiting, and logging
- Never forward the employee API key to the upstream; always overwrite with the upstream credential

## Request Flow

1. Client calls `/proxy/:api_name/...`
2. Middleware first checks `Authorization: Bearer ...`
3. If Bearer auth is missing, the middleware loads the upstream by `api_name`
4. If that upstream has `allow_native_client_auth=true`, the middleware reads the employee API key from the upstream's native auth location
5. The gateway validates the employee API key as usual
6. The proxy director rewrites auth on the outbound request using the upstream's configured credential

## UI

- Add a toggle to the upstream editor for `allow_native_client_auth`
- Explain that enabling it lets clients send the employee API key in the upstream's native auth location while Bearer auth still works

## Verification

- Middleware tests for Bearer, query, and header auth flows
- Tests that Bearer takes priority over native compatibility
- Proxy director tests that query/header client credentials are overwritten before forwarding
