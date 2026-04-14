package proxy

import (
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * 1e9,
		TLSHandshakeTimeout:   10 * 1e9,
		ExpectContinueTimeout: 1 * 1e9,
	}
}

func NewTransportWithProxy(proxyURL string) (*http.Transport, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "socks5", "socks5h":
		return newSocks5Transport(u)
	default:
		return &http.Transport{
			Proxy:                 http.ProxyURL(u),
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * 1e9,
			TLSHandshakeTimeout:   10 * 1e9,
			ExpectContinueTimeout: 1 * 1e9,
		}, nil
	}
}

func newSocks5Transport(u *url.URL) (*http.Transport, error) {
	auth := &proxy.Auth{}
	if u.User != nil {
		auth.User = u.User.Username()
		auth.Password, _ = u.User.Password()
	} else {
		auth = nil
	}
	dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		Dial:                  dialer.Dial,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * 1e9,
		TLSHandshakeTimeout:   10 * 1e9,
		ExpectContinueTimeout: 1 * 1e9,
	}, nil
}
