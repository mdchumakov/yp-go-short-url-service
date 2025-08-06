package urlShortener

import (
	"context"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

type linkShortenerService struct {
	repository repository.URLRepository
}

func NewLinkShortenerService(repository repository.URLRepository) service.LinkShortenerService {
	return &linkShortenerService{
		repository: repository,
	}
}

func (s *linkShortenerService) ShortURL(ctx context.Context, longURL string) (string, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	logger.Infow("Starting URL shortening process",
		"long_url", longURL,
		"request_id", requestID,
	)

	shortURLFromStorage, err := s.extractShortURLIfExists(ctx, longURL)
	if err != nil {
		logger.Errorw("Failed to extract short URL from storage",
			"error", err,
			"long_url", longURL,
			"request_id", requestID,
		)
		return "", err
	}

	if shortURLFromStorage != nil {
		logger.Infow("Short URL already exists in storage",
			"short_url", *shortURLFromStorage,
			"long_url", longURL,
			"request_id", requestID,
		)
		return *shortURLFromStorage, nil
	}

	logger.Infow("Short URL not found in storage, generating new one",
		"long_url", longURL,
		"request_id", requestID,
	)

	shortURL := shortenURLBase62(longURL)
	logger.Debugw("Generated short URL",
		"short_url", shortURL,
		"request_id", requestID,
	)

	newURL := model.URLsModel{
		ShortURL: shortURL,
		LongURL:  longURL,
	}

	err = s.saveShortURLToStorage(ctx, &newURL)
	if err != nil {
		logger.Errorw("Failed to save short URL to storage",
			"error", err,
			"short_url", shortURL,
			"long_url", longURL,
			"request_id", requestID,
		)
		return "", err
	}

	logger.Infow("Short URL successfully saved to storage",
		"short_url", newURL.ShortURL,
		"long_url", newURL.LongURL,
		"id", newURL.ID,
		"request_id", requestID,
	)

	return newURL.ShortURL, nil
}

func (s *linkShortenerService) extractShortURLIfExists(ctx context.Context, longURL string) (*string, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	urlResponse, err := s.repository.GetByLongURL(ctx, longURL)
	if err != nil {
		if repository.IsNotFoundError(err) {
			logger.Debugw("URL not found in storage",
				"long_url", longURL,
				"request_id", requestID)
			return nil, nil
		} else {
			logger.Errorw("Database error while searching for URL",
				"error", err,
				"long_url", longURL,
				"request_id", requestID)
			return nil, err
		}
	}

	logger.Debugw("Found existing short URL",
		"short_url", urlResponse.ShortURL,
		"long_url", longURL,
		"request_id", requestID)
	return &urlResponse.ShortURL, nil
}

func (s *linkShortenerService) saveShortURLToStorage(ctx context.Context, url *model.URLsModel) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	err := s.repository.Create(ctx, url)
	if err != nil {
		logger.Errorw("Failed to create URL record in database",
			"error", err,
			"short_url", url.ShortURL,
			"long_url", url.LongURL,
			"request_id", requestID)
		return err
	}

	logger.Debugw("URL record created successfully in database",
		"short_url", url.ShortURL,
		"long_url", url.LongURL,
		"request_id", requestID)
	return nil
}
