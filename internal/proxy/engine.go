package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"time"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
)

type Engine struct {
	upstreamSvc *service.UpstreamService
	apiKeySvc   *service.ApiKeyService
	logger      *service.RequestLogger
	transport   *http.Transport
}

func NewEngine(upstreamSvc *service.UpstreamService, apiKeySvc *service.ApiKeyService, logger *service.RequestLogger) *Engine {
	return &Engine{
		upstreamSvc: upstreamSvc,
		apiKeySvc:   apiKeySvc,
		logger:      logger,
		transport:   NewTransport(),
	}
}

func (e *Engine) ResolveUpstream(name string) (*model.Upstream, error) {
	up, err := e.upstreamSvc.GetActiveByName(name)
	if err != nil {
		return nil, err
	}
	return up, nil
}

func (e *Engine) BuildProxy(upstream *model.Upstream, meta *MetaCarrier) (*httputil.ReverseProxy, error) {
	if upstream == nil {
		return nil, errors.New("upstream is nil")
	}
	director, err := BuildDirector(upstream, meta)
	if err != nil {
		return nil, err
	}
	transport, err := e.pickTransport(upstream)
	if err != nil {
		return nil, err
	}
	p := &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		FlushInterval:  -1,
		ModifyResponse: ModifyResponse(meta),
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			if meta != nil {
				meta.EndedAt = time.Now()
				meta.ErrorMsg = err.Error()
			}
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte(`{"error":"upstream request failed"}`))
		},
	}
	return p, nil
}

func (e *Engine) pickTransport(upstream *model.Upstream) (http.RoundTripper, error) {
	if upstream.ProxyURL == "" {
		return e.transport, nil
	}
	return NewTransportWithProxy(upstream.ProxyURL)
}

func (e *Engine) AfterProxy(log *model.RequestLog) {
	HandlePostLog(e.logger, e.apiKeySvc, log)
}
