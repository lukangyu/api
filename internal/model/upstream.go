package model

import "time"

type Upstream struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	Name                  string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	DisplayName           string    `gorm:"size:128;not null" json:"display_name"`
	BaseURL               string    `gorm:"size:512;not null" json:"base_url"`
	AuthType              string    `gorm:"size:16;not null;default:'none'" json:"auth_type"`
	AuthKey               string    `gorm:"size:128;not null;default:''" json:"auth_key"`
	AuthValue             string    `gorm:"size:512;not null;default:''" json:"auth_value"`
	AllowNativeClientAuth bool      `gorm:"not null;default:false" json:"allow_native_client_auth"`
	TimeoutSeconds        int       `gorm:"not null;default:120" json:"timeout_seconds"`
	ProxyURL              string    `gorm:"size:512;not null;default:''" json:"proxy_url"`
	StripPrefix           bool      `gorm:"not null;default:true" json:"strip_prefix"`
	ExtraHeaders          string    `gorm:"type:text;not null;default:'{}'" json:"extra_headers"`
	IsActive              bool      `gorm:"not null;default:true" json:"is_active"`
	Description           string    `gorm:"type:text;not null;default:''" json:"description"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
