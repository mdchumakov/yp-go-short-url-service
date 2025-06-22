package service

import (
	"fmt"
	"gorm.io/gorm"
	"yp-go-short-url-service/internal/model"
)

type LinkShortener interface {
	ShortenURL(longURL string) (string, error)
}

type linkShortenerService struct {
	db *gorm.DB
}

func NewLinkShortenerService(db *gorm.DB) LinkShortener {
	return &linkShortenerService{db: db}
}

func (s *linkShortenerService) ShortenURL(longURL string) (string, error) {
	var urlModel model.URL

	s.db.First(&urlModel, "long_url = ?", longURL)
	if urlModel.ID != 0 {
		return urlModel.ShortURL, nil
	}

	shortURL := shortenURLBase62(longURL)
	fmt.Println(shortURL)

	s.db.First(&urlModel, "short_url = ?", shortURL)
	if urlModel.ID != 0 {
		resolveConflicts(s.db, &shortURL)
	}

	s.db.Create(&model.URL{ShortURL: shortURL, LongURL: longURL})
	return shortURL, nil
}

func resolveConflicts(db *gorm.DB, shortURL *string) {
	var urlModel model.URL
	for {
		db.First(&urlModel, "short_url = ?", shortURL)
		if urlModel.ID == 0 {
			break
		}
		*shortURL = shortenURLBase62(*shortURL + "1")
	}
}
