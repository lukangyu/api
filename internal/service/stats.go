package service

import (
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

type Overview struct {
	TotalRequests int64 `json:"total_requests"`
	ActiveUsers   int64 `json:"active_users"`
	ActiveAPIKeys int64 `json:"active_api_keys"`
	Upstreams     int64 `json:"upstreams"`
}

func (s *StatsService) Overview() (Overview, error) {
	var out Overview
	if err := s.db.Model(&model.RequestLog{}).Count(&out.TotalRequests).Error; err != nil {
		return out, err
	}
	if err := s.db.Model(&model.User{}).Where("is_active = ?", true).Count(&out.ActiveUsers).Error; err != nil {
		return out, err
	}
	if err := s.db.Model(&model.ApiKey{}).Where("is_active = ?", true).Count(&out.ActiveAPIKeys).Error; err != nil {
		return out, err
	}
	if err := s.db.Model(&model.Upstream{}).Where("is_active = ?", true).Count(&out.Upstreams).Error; err != nil {
		return out, err
	}
	return out, nil
}

type DailyPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (s *StatsService) Daily(days int) ([]DailyPoint, error) {
	if days <= 0 || days > 365 {
		days = 7
	}
	start := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
	type row struct {
		Date  string
		Count int64
	}
	var rows []row
	err := s.db.Table("request_logs").
		Select("date(created_at) as date, count(*) as count").
		Where("date(created_at) >= ?", start).
		Group("date(created_at)").
		Order("date(created_at) asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]DailyPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, DailyPoint{Date: r.Date, Count: r.Count})
	}
	return out, nil
}
