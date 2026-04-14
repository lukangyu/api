package database

import (
	"api_zhuanfa/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB, username, password string) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		Username:     username,
		PasswordHash: string(h),
		DisplayName:  "管理员",
		Role:         "admin",
		IsActive:     true,
	}
	return db.Create(&admin).Error
}
