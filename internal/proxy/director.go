package proxy

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"api_zhuanfa/internal/model"
)

func BuildDirector(upstream *model.Upstream) (func(*http.Request), error) {
	target, err := url.Parse(upstream.BaseURL)
	if err != nil {
		return nil, err
	}
	basePath := strings.TrimRight(target.Path, "/")

	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		path := req.URL.Path
		if upstream.StripPrefix {
			parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)
			if len(parts) >= 2 && parts[0] == "proxy" {
				if len(parts) == 2 {
					path = "/"
				} else {
					path = "/" + parts[2]
				}
			}
		}
		req.URL.Path = singleJoiningSlash(basePath, path)

		applyUpstreamAuth(req, upstream)
		applyExtraHeaders(req, upstream.ExtraHeaders)

		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}, nil
}

func applyUpstreamAuth(req *http.Request, up *model.Upstream) {
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

func applyExtraHeaders(req *http.Request, raw string) {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "{}" {
		return
	}
	m := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return
	}
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
