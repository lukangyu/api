package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

type ApiKeyService struct {
	db *gorm.DB
}

func NewApiKeyService(db *gorm.DB) *ApiKeyService {
	return &ApiKeyService{db: db}
}

func (s *ApiKeyService) Generate(userID uint, name string, requestLimit int64, expiresAt *time.Time, allowedUpstreams string) (string, *model.ApiKey, error) {
	plain, hash, prefix, err := newAPIKey()
	if err != nil {
		return "", nil, err
	}

	normalized, err := NormalizeAllowedUpstreams(allowedUpstreams)
	if err != nil {
		return "", nil, err
	}

	entity := &model.ApiKey{
		UserID:           userID,
		KeyHash:          hash,
		KeyPrefix:        prefix,
		Name:             name,
		RequestLimit:     requestLimit,
		ExpiresAt:        expiresAt,
		AllowedUpstreams: normalized,
		IsActive:         true,
	}

	if err := s.db.Create(entity).Error; err != nil {
		return "", nil, err
	}

	return plain, entity, nil
}

func (s *ApiKeyService) List(page, pageSize int) ([]model.ApiKey, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	var total int64
	if err := s.db.Model(&model.ApiKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ApiKey
	err := s.db.Preload("User").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (s *ApiKeyService) Revoke(id uint) error {
	return s.db.Model(&model.ApiKey{}).Where("id = ?", id).Update("is_active", false).Error
}

func (s *ApiKeyService) Update(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	if raw, ok := updates["allowed_upstream_ids"]; ok {
		ids, err := CoerceAllowedUpstreamIDs(raw)
		if err != nil {
			return err
		}
		updates["allowed_upstreams"] = JoinAllowedUpstreamIDs(ids)
		delete(updates, "allowed_upstream_ids")
	}
	if raw, ok := updates["allowed_upstreams"]; ok {
		v, ok := raw.(string)
		if !ok {
			return errors.New("allowed_upstreams must be a string")
		}
		normalized, err := NormalizeAllowedUpstreams(v)
		if err != nil {
			return err
		}
		updates["allowed_upstreams"] = normalized
	}
	return s.db.Model(&model.ApiKey{}).Where("id = ?", id).Updates(updates).Error
}

func (s *ApiKeyService) Validate(plainKey string) (*model.ApiKey, *model.User, error) {
	h := HashKey(plainKey)
	var key model.ApiKey
	if err := s.db.Where("key_hash = ?", h).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid api key")
		}
		return nil, nil, err
	}
	if !key.IsActive {
		return nil, nil, errors.New("api key disabled")
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, nil, errors.New("api key expired")
	}
	if key.RequestLimit > 0 && key.RequestCount >= key.RequestLimit {
		return nil, nil, errors.New("api key quota exceeded")
	}

	var user model.User
	if err := s.db.First(&user, key.UserID).Error; err != nil {
		return nil, nil, err
	}
	if !user.IsActive {
		return nil, nil, errors.New("user disabled")
	}
	return &key, &user, nil
}

func (s *ApiKeyService) CheckUpstreamAllowed(key *model.ApiKey, upstreamID uint) bool {
	ids := ParseAllowedUpstreamIDs(key.AllowedUpstreams)
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == upstreamID {
			return true
		}
	}
	return false
}

func (s *ApiKeyService) IncRequestCount(id uint) {
	_ = s.db.Model(&model.ApiKey{}).Where("id = ?", id).UpdateColumn("request_count", gorm.Expr("request_count + 1")).Error
}

func newAPIKey() (plain, hash, prefix string, err error) {
	buf := make([]byte, 18)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	raw := hex.EncodeToString(buf)
	plain = "sk-" + raw
	hash = HashKey(plain)
	if len(plain) >= 11 {
		prefix = plain[:11]
	} else {
		prefix = plain
	}
	return
}

func HashKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func ParseAllowedUpstreamIDs(raw string) []uint {
	parts := strings.Split(raw, ",")
	seen := make(map[uint]struct{}, len(parts))
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id64, err := strconv.ParseUint(part, 10, 64)
		if err != nil || id64 == 0 {
			continue
		}
		id := uint(id64)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func JoinAllowedUpstreamIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	seen := make(map[uint]struct{}, len(ids))
	unique := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return ""
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })
	out := make([]string, 0, len(unique))
	for _, id := range unique {
		out = append(out, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(out, ",")
}

func NormalizeAllowedUpstreams(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	result := JoinAllowedUpstreamIDs(ParseAllowedUpstreamIDs(raw))
	if result == "" {
		return "", fmt.Errorf("allowed_upstreams 包含无效值: %q", trimmed)
	}
	return result, nil
}

func CoerceAllowedUpstreamIDs(raw interface{}) ([]uint, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []uint:
		return append([]uint(nil), v...), nil
	case []int:
		out := make([]uint, 0, len(v))
		for _, id := range v {
			if id < 0 {
				return nil, fmt.Errorf("allowed_upstream_ids contains invalid value %d", id)
			}
			out = append(out, uint(id))
		}
		return out, nil
	case []string:
		out := make([]uint, 0, len(v))
		for _, item := range v {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			id64, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("allowed_upstream_ids contains invalid value %q", item)
			}
			out = append(out, uint(id64))
		}
		return out, nil
	case []interface{}:
		out := make([]uint, 0, len(v))
		for _, item := range v {
			id, err := coerceAllowedUpstreamID(item)
			if err != nil {
				return nil, err
			}
			out = append(out, id)
		}
		return out, nil
	default:
		return nil, errors.New("allowed_upstream_ids must be an array")
	}
}

func coerceAllowedUpstreamID(raw interface{}) (uint, error) {
	switch v := raw.(type) {
	case uint:
		if v == 0 {
			return 0, errors.New("allowed_upstream_ids contains invalid value 0")
		}
		return v, nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("allowed_upstream_ids contains invalid value %d", v)
		}
		return uint(v), nil
	case float64:
		if v <= 0 || math.Trunc(v) != v {
			return 0, fmt.Errorf("allowed_upstream_ids contains invalid value %v", v)
		}
		return uint(v), nil
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return 0, errors.New("allowed_upstream_ids contains empty value")
		}
		id64, err := strconv.ParseUint(v, 10, 64)
		if err != nil || id64 == 0 {
			return 0, fmt.Errorf("allowed_upstream_ids contains invalid value %q", v)
		}
		return uint(id64), nil
	default:
		return 0, fmt.Errorf("allowed_upstream_ids contains unsupported value %T", raw)
	}
}
