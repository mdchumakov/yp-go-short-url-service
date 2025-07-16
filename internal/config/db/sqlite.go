package db

import (
	"errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"yp-go-short-url-service/internal/model"

	"yp-go-short-url-service/internal/utils"
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

func SetupDB(dbPath string, fileStoragePath string, log *zap.SugaredLogger) (*gorm.DB, error) {
	sqliteDB, err := InitSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	if isFileExists := utils.CheckFileExists(fileStoragePath); !isFileExists {
		log.Warnf("File storage path %s does not exist, skipping URL insertion", fileStoragePath)
		urlFromDB, err := extractURLSDataFromDB(sqliteDB, log)
		if err != nil {
			log.Error("Error extracting URLs from SQLite DB", zap.Error(err))
			return nil, err
		}
		if err := SaveURLSDataToFileStorage(fileStoragePath, urlFromDB, log); err != nil {
			log.Error("Error saving URLs to file storage", zap.Error(err))
			return nil, err
		}
		log.Info("File storage path does not exist, URLs extracted from DB and saved to file storage")
	} else {
		urlData, err := ExtractURLSDataFromFileStorage(fileStoragePath, log)
		if err != nil {
			return nil, err
		}
		if err := loadURLData2DB(sqliteDB, urlData, log); err != nil {
			log.Error("Error loading data into SQLite DB", zap.Error(err))
			return nil, err
		}
	}

	log.Info("Successfully initialized SQLite DB and inserted URLs from file storage")
	return sqliteDB, nil
}

func extractURLSDataFromDB(db *gorm.DB, log *zap.SugaredLogger) ([]model.URL, error) {
	var urls []model.URL
	if err := db.Find(&urls).Error; err != nil {
		log.Error("Error extracting URLs from SQLite DB", zap.Error(err))
		return nil, err
	}
	log.Infof("Successfully extracted %d URLs from SQLite DB", len(urls))
	return urls, nil
}

func loadURLData2DB(db *gorm.DB, urls []URL, log *zap.SugaredLogger) error {
	for _, url := range urls {
		if err := db.Create(&url).Error; err != nil {
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
				return err
			}
		}
	}
	return nil
}
