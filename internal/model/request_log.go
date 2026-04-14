package model

import "time"

type RequestLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ApiKeyID      uint      `gorm:"index;not null" json:"api_key_id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	UpstreamID    uint      `gorm:"index;not null" json:"upstream_id"`
	Method        string    `gorm:"size:16;not null" json:"method"`
	Path          string    `gorm:"size:1024;not null" json:"path"`
	UpstreamPath  string    `gorm:"size:1024;not null;default:''" json:"upstream_path"`
	StatusCode    int       `gorm:"not null;default:0" json:"status_code"`
	RequestBytes  int64     `gorm:"not null;default:0" json:"request_bytes"`
	ResponseBytes int64     `gorm:"not null;default:0" json:"response_bytes"`
	LatencyMs     int64     `gorm:"not null;default:0" json:"latency_ms"`
	ClientIP      string    `gorm:"size:64;not null;default:''" json:"client_ip"`
	ErrorMessage  string    `gorm:"type:text;not null;default:''" json:"error_message"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}
