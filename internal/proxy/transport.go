package proxy

import "net/http"

func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * 1e9,
		TLSHandshakeTimeout:   10 * 1e9,
		ExpectContinueTimeout: 1 * 1e9,
	}
}
