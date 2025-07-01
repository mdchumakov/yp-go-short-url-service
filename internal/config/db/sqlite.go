package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"yp-go-short-url-service/internal/model"
)

type SQLiteSettings struct {
	SQLiteDBPath string `envconfig:"SQLITE_DB_PATH" default:"db/test.db" required:"true"`
}

func InitSQLiteDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&model.URL{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
