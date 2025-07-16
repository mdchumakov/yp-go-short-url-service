package db

import (
	"errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"

	"yp-go-short-url-service/internal/model"
)

type SQLiteSettings struct {
	SQLiteDBPath string `envconfig:"SQLITE_DB_PATH" default:"db/test.db" required:"true"`
}

func SetupDB(dbPath string, fileStoragePath string, log *zap.SugaredLogger) (*gorm.DB, error) {
	sqliteDB, err := InitSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	urlData, err := ExtractURLSDataFromFileStorage(fileStoragePath, log)
	if err != nil {
		return nil, err
	}

	for _, url := range urlData {
		if err := sqliteDB.Create(&url).Error; err != nil {
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				log.Warnf("URL with short URL %s already exists in the database, skipping insertion", url.ShortURL)
				continue
			case strings.Contains(err.Error(), "UNIQUE constraint"):
				log.Warnf("URL with short URL %s already exists", url.ShortURL)
				continue
			case strings.Contains(err.Error(), "UNIQUE constraint failed"):
				log.Warnf("URL with short URL %s already exists in the database, skipping insertion", url.ShortURL)
				continue
			default:
				log.Error("Error inserting URL into SQLite DB", zap.Error(err), zap.String("short_url", url.ShortURL), zap.String("long_url", url.LongURL))
				return nil, err
			}
		}
	}

	log.Info("Successfully initialized SQLite DB and inserted URLs from file storage")
	return sqliteDB, nil
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
