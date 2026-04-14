package model

type DailyStats struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Date          string `gorm:"size:10;not null;index:idx_daily_unique,unique" json:"date"`
	UserID        uint   `gorm:"not null;index:idx_daily_unique,unique" json:"user_id"`
	UpstreamID    uint   `gorm:"not null;index:idx_daily_unique,unique" json:"upstream_id"`
	RequestCount  int64  `gorm:"not null;default:0" json:"request_count"`
	TotalBytesIn  int64  `gorm:"not null;default:0" json:"total_bytes_in"`
	TotalBytesOut int64  `gorm:"not null;default:0" json:"total_bytes_out"`
	AvgLatencyMs  int64  `gorm:"not null;default:0" json:"avg_latency_ms"`
	ErrorCount    int64  `gorm:"not null;default:0" json:"error_count"`
}
