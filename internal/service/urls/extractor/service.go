package extractor

import (
	"context"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func NewLinkExtractorService(
	urlRepository repository.URLRepositoryReader,
	userURLsRepository repository.UserURLsRepositoryReader,
) service.LinkExtractorService {
	return &linkExtractorService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
	}
}

type linkExtractorService struct {
	urlRepository      repository.URLRepositoryReader
	userURLsRepository repository.UserURLsRepositoryReader
}

func (s *linkExtractorService) ExtractUserURLs(ctx context.Context, userID string) ([]*model.URLsModel, error) {
	logger := middleware.GetLogger(ctx)
	logger.Infow("Starting user URLs extraction process",
		"user_id", userID,
		"request_id", middleware.ExtractRequestID(ctx),
	)

	urls, err := s.userURLsRepository.GetByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to extract user URLs from storage",
			"error", err,
			"user_id", userID,
			"request_id", middleware.ExtractRequestID(ctx),
		)
		return nil, err
	}

	logger.Infow("Successfully extracted user URLs from storage",
		"user_id", userID,
		"urls_count", len(urls),
		"request_id", middleware.ExtractRequestID(ctx),
	)

	return urls, nil
}

func (s *linkExtractorService) ExtractLongURL(ctx context.Context, shortURL string) (string, error) {
	logger := middleware.GetLogger(ctx)
	logger.Infow("Starting URL extraction process",
		"short_url", shortURL,
		"request_id", middleware.ExtractRequestID(ctx),
	)

	url, err := s.urlRepository.GetByShortURL(ctx, shortURL)
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
