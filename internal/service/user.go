package service

import (
	"errors"
	"strings"

	"api_zhuanfa/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	var total int64
	if err := s.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := s.db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (s *UserService) Create(username, password, displayName, role string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" {
		role = "user"
	}
	h, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     username,
		PasswordHash: h,
		DisplayName:  displayName,
		Role:         role,
		IsActive:     true,
	}
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Update(id uint, displayName, role string, active *bool) error {
	updates := map[string]interface{}{}
	if displayName != "" {
		updates["display_name"] = displayName
	}
	if role == "admin" || role == "user" {
		updates["role"] = role
	}
	if active != nil {
		updates["is_active"] = *active
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) Deactivate(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error
}

func (s *UserService) ResetPassword(id uint, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("密码不能为空")
	}
	h, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", h).Error
}
