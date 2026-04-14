package proxy

import (
	"net/http"
	"time"

	"api_zhuanfa/internal/model"
	"api_zhuanfa/internal/service"
)

type MetaCarrier struct {
	StatusCode   int
	BodyBytes    int64
	StartedAt    time.Time
	EndedAt      time.Time
	ErrorMsg     string
	UpstreamPath string
}

func ModifyResponse(meta *MetaCarrier) func(*http.Response) error {
	return func(resp *http.Response) error {
		if meta != nil {
			meta.StatusCode = resp.StatusCode
			meta.BodyBytes = resp.ContentLength
			meta.EndedAt = time.Now()
		}
		return nil
	}
}

func BuildLogEntry(meta *MetaCarrier, originalPath string, req *http.Request, apiKeyID, userID, upstreamID uint) *model.RequestLog {
	latency := int64(0)
	if meta != nil && !meta.StartedAt.IsZero() && !meta.EndedAt.IsZero() {
		latency = meta.EndedAt.Sub(meta.StartedAt).Milliseconds()
	}
	requestBytes := req.ContentLength
	if requestBytes < 0 {
		requestBytes = 0
	}
	statusCode := 0
	responseBytes := int64(0)
	errorMsg := ""
	upstreamPath := ""
	if meta != nil {
		statusCode = meta.StatusCode
		if meta.BodyBytes > 0 {
			responseBytes = meta.BodyBytes
		}
		errorMsg = meta.ErrorMsg
		upstreamPath = meta.UpstreamPath
	}
	return &model.RequestLog{
		ApiKeyID:      apiKeyID,
		UserID:        userID,
		UpstreamID:    upstreamID,
		Method:        req.Method,
		Path:          originalPath,
		UpstreamPath:  upstreamPath,
		StatusCode:    statusCode,
		RequestBytes:  requestBytes,
		ResponseBytes: responseBytes,
		LatencyMs:     latency,
		ClientIP:      req.RemoteAddr,
		ErrorMessage:  errorMsg,
	}
}

func HandlePostLog(logger *service.RequestLogger, keySvc *service.ApiKeyService, log *model.RequestLog) {
	if log == nil {
		return
	}
	logger.Log(log)
	keySvc.IncRequestCount(log.ApiKeyID)
}
