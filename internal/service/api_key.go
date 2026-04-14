package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

	entity := &model.ApiKey{
		UserID:           userID,
		KeyHash:          hash,
		KeyPrefix:        prefix,
		Name:             name,
		RequestLimit:     requestLimit,
		ExpiresAt:        expiresAt,
		AllowedUpstreams: strings.TrimSpace(allowedUpstreams),
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
	allowRaw := strings.TrimSpace(key.AllowedUpstreams)
	if allowRaw == "" {
		return true
	}
	needle := strconv.FormatUint(uint64(upstreamID), 10)
	for _, p := range strings.Split(allowRaw, ",") {
		if strings.TrimSpace(p) == needle {
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
