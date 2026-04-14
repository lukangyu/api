package service

import (
	"errors"
	"strings"
	"sync"
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

type UpstreamService struct {
	db       *gorm.DB
	ttl      time.Duration
	mu       sync.RWMutex
	byName   map[string]model.Upstream
	lastLoad time.Time
}

func NewUpstreamService(db *gorm.DB, cacheTTLSeconds int) *UpstreamService {
	if cacheTTLSeconds <= 0 {
		cacheTTLSeconds = 30
	}
	return &UpstreamService{
		db:     db,
		ttl:    time.Duration(cacheTTLSeconds) * time.Second,
		byName: make(map[string]model.Upstream),
	}
}

func (s *UpstreamService) List() ([]model.Upstream, error) {
	var rows []model.Upstream
	return rows, s.db.Order("id desc").Find(&rows).Error
}

func (s *UpstreamService) Create(in *model.Upstream) error {
	if in == nil {
		return errors.New("empty upstream")
	}
	in.Name = strings.TrimSpace(in.Name)
	in.BaseURL = strings.TrimSpace(in.BaseURL)
	if in.Name == "" || in.BaseURL == "" {
		return errors.New("name/base_url 不能为空")
	}
	if in.AuthType == "" {
		in.AuthType = "none"
	}
	if in.TimeoutSeconds <= 0 {
		in.TimeoutSeconds = 120
	}
	if in.ExtraHeaders == "" {
		in.ExtraHeaders = "{}"
	}
	if err := s.db.Create(in).Error; err != nil {
		return err
	}
	s.Invalidate()
	return nil
}

func (s *UpstreamService) Update(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	if err := s.db.Model(&model.Upstream{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	s.Invalidate()
	return nil
}

func (s *UpstreamService) Delete(id uint) error {
	if err := s.db.Delete(&model.Upstream{}, id).Error; err != nil {
		return err
	}
	s.Invalidate()
	return nil
}

func (s *UpstreamService) GetActiveByName(name string) (*model.Upstream, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("empty upstream name")
	}
	if err := s.ensureCache(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	row, ok := s.byName[name]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cpy := row
	return &cpy, nil
}

func (s *UpstreamService) Invalidate() {
	s.mu.Lock()
	s.lastLoad = time.Time{}
	s.mu.Unlock()
}

func (s *UpstreamService) ensureCache() error {
	s.mu.RLock()
	fresh := !s.lastLoad.IsZero() && time.Since(s.lastLoad) < s.ttl
	s.mu.RUnlock()
	if fresh {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.lastLoad.IsZero() && time.Since(s.lastLoad) < s.ttl {
		return nil
	}
	var rows []model.Upstream
	if err := s.db.Where("is_active = ?", true).Find(&rows).Error; err != nil {
		return err
	}
	m := make(map[string]model.Upstream, len(rows))
	for _, r := range rows {
		m[r.Name] = r
	}
	s.byName = m
	s.lastLoad = time.Now()
	return nil
}
