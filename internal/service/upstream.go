package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

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
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.BaseURL = strings.TrimSpace(in.BaseURL)
	in.AuthType = strings.TrimSpace(in.AuthType)
	in.AuthKey = strings.TrimSpace(in.AuthKey)

	if in.Name == "" {
		return errors.New("name 不能为空")
	}
	if len(in.Name) > 64 || !nameRe.MatchString(in.Name) {
		return errors.New("name 只能包含字母、数字、下划线和连字符，最长64位")
	}
	if in.DisplayName == "" {
		return errors.New("display_name 不能为空")
	}
	if in.BaseURL == "" {
		return errors.New("base_url 不能为空")
	}
	if err := validateBaseURL(in.BaseURL); err != nil {
		return err
	}
	if in.AuthType == "" {
		in.AuthType = "none"
	}
	if err := validateAuthType(in.AuthType); err != nil {
		return err
	}
	if (in.AuthType == "header" || in.AuthType == "query") && in.AuthKey == "" {
		return fmt.Errorf("auth_type 为 %s 时 auth_key 不能为空", in.AuthType)
	}
	if in.TimeoutSeconds <= 0 {
		in.TimeoutSeconds = 120
	}
	if in.ExtraHeaders == "" {
		in.ExtraHeaders = "{}"
	}
	if err := validateExtraHeaders(in.ExtraHeaders); err != nil {
		return err
	}
	if err := s.db.Create(in).Error; err != nil {
		return err
	}
	s.Invalidate()
	return nil
}

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("base_url 格式无效: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("base_url 必须以 http:// 或 https:// 开头")
	}
	if u.Host == "" {
		return errors.New("base_url 缺少主机名")
	}
	return nil
}

func validateAuthType(t string) error {
	switch t {
	case "none", "bearer", "header", "query":
		return nil
	default:
		return fmt.Errorf("auth_type 无效，可选: none, bearer, header, query")
	}
}

func validateExtraHeaders(raw string) error {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "{}" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return fmt.Errorf("extra_headers 必须是合法的 JSON 对象: %v", err)
	}
	return nil
}

var upstreamAllowedFields = map[string]bool{
	"name": true, "display_name": true, "base_url": true,
	"auth_type": true, "auth_key": true, "auth_value": true,
	"timeout_seconds": true, "strip_prefix": true, "extra_headers": true,
	"is_active": true, "description": true,
}

func (s *UpstreamService) Update(id uint, raw map[string]interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	filtered := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		if upstreamAllowedFields[k] {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	if v, ok := strVal(filtered, "name"); ok {
		v = strings.TrimSpace(v)
		if v == "" {
			return errors.New("name 不能为空")
		}
		if len(v) > 64 || !nameRe.MatchString(v) {
			return errors.New("name 只能包含字母、数字、下划线和连字符，最长64位")
		}
		filtered["name"] = v
	}
	if v, ok := strVal(filtered, "display_name"); ok {
		if strings.TrimSpace(v) == "" {
			return errors.New("display_name 不能为空")
		}
	}
	if v, ok := strVal(filtered, "base_url"); ok {
		v = strings.TrimSpace(v)
		if v == "" {
			return errors.New("base_url 不能为空")
		}
		if err := validateBaseURL(v); err != nil {
			return err
		}
		filtered["base_url"] = v
	}

	authType, atOk := strVal(filtered, "auth_type")
	authKey, akOk := strVal(filtered, "auth_key")
	if atOk {
		authType = strings.TrimSpace(authType)
		if err := validateAuthType(authType); err != nil {
			return err
		}
		filtered["auth_type"] = authType
	}
	if atOk && (authType == "header" || authType == "query") {
		key := strings.TrimSpace(authKey)
		if akOk && key == "" {
			return fmt.Errorf("auth_type 为 %s 时 auth_key 不能为空", authType)
		}
		if !akOk {
			var existing model.Upstream
			if err := s.db.Select("auth_key").Where("id = ?", id).First(&existing).Error; err != nil {
				return err
			}
			if strings.TrimSpace(existing.AuthKey) == "" {
				return fmt.Errorf("auth_type 为 %s 时 auth_key 不能为空", authType)
			}
		}
	}

	if v, ok := strVal(filtered, "extra_headers"); ok {
		if err := validateExtraHeaders(v); err != nil {
			return err
		}
	}

	if err := s.db.Model(&model.Upstream{}).Where("id = ?", id).Updates(filtered).Error; err != nil {
		return err
	}
	s.Invalidate()
	return nil
}

func strVal(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
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
