package model

import "time"

type ApiKey struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"index;not null" json:"user_id"`
	KeyHash          string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	KeyPrefix        string     `gorm:"size:16;not null" json:"key_prefix"`
	Name             string     `gorm:"size:128;not null;default:''" json:"name"`
	IsActive         bool       `gorm:"not null;default:true" json:"is_active"`
	RequestLimit     int64      `gorm:"not null;default:0" json:"request_limit"`
	RequestCount     int64      `gorm:"not null;default:0" json:"request_count"`
	ExpiresAt        *time.Time `json:"expires_at"`
	AllowedUpstreams string     `gorm:"type:text;not null;default:''" json:"allowed_upstreams"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
