package database

import (
	"fmt"
	"os"
	"path/filepath"

	"api_zhuanfa/internal/model"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Init(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	_ = db.Exec("PRAGMA journal_mode=DELETE;").Error
	_ = db.Exec("PRAGMA busy_timeout=5000;").Error
	_ = db.Exec("PRAGMA synchronous=NORMAL;").Error

	if err := db.AutoMigrate(
		&model.User{},
		&model.ApiKey{},
		&model.Upstream{},
		&model.RequestLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}
