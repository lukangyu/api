package service

import (
	"errors"
	"strconv"
	"time"

	"api_zhuanfa/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db            *gorm.DB
	jwtSecret     string
	jwtExpireHour int
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpireHour int) *AuthService {
	if jwtExpireHour <= 0 {
		jwtExpireHour = 24
	}
	return &AuthService{db: db, jwtSecret: jwtSecret, jwtExpireHour: jwtExpireHour}
}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	var user model.User
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户名或密码错误")
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	token, err := s.GenerateToken(&user)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func (s *AuthService) GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  strconv.FormatUint(uint64(user.ID), 10),
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(time.Duration(s.jwtExpireHour) * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.jwtSecret))
}

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}
