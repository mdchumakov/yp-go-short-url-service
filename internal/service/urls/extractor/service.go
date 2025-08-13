package extractor

import (
	"context"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

type linkExtractorService struct {
	repository repository.URLRepositoryReader
}

func NewLinkExtractorService(repository repository.URLRepositoryReader) service.LinkExtractorService {
	return &linkExtractorService{
		repository: repository,
	}
}

func (s *linkExtractorService) ExtractLongURL(ctx context.Context, shortURL string) (string, error) {
	logger := middleware.GetLogger(ctx)
	logger.Infow("Starting URL extraction process",
		"short_url", shortURL,
		"request_id", middleware.ExtractRequestID(ctx),
	)

	url, err := s.repository.GetByShortURL(ctx, shortURL)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return "", nil
		}
		logger.Errorw("Failed to extract long URL from storage",
			"error", err,
			"short_url", shortURL,
			"request_id", middleware.ExtractRequestID(ctx),
		)
		return "", err
	}

	if url == nil {
		logger.Warnw("Short URL not found in storage",
			"short_url", shortURL,
			"request_id", middleware.ExtractRequestID(ctx),
		)
		return "", nil
	}

	logger.Infow("Successfully extracted long URL from storage",
		"long_url", url.LongURL,
		"short_url", shortURL,
		"request_id", middleware.ExtractRequestID(ctx),
	)

	return url.LongURL, nil
}
