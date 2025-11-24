package extractor

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/observer/audit"
	baseObserver "yp-go-short-url-service/internal/observer/base"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func NewLinkExtractorService(
	urlRepository repository.URLRepositoryReader,
	userURLsRepository repository.UserURLsRepositoryReader,
	eventBus baseObserver.Subject[audit.Event],
) service.URLExtractorService {
	return &linkExtractorService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
		eventBus:           eventBus,
	}
}

type linkExtractorService struct {
	urlRepository      repository.URLRepositoryReader
	userURLsRepository repository.UserURLsRepositoryReader
	eventBus           baseObserver.Subject[audit.Event]
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
	requestID := middleware.ExtractRequestID(ctx)

	logger.Infow("Starting URL extraction process",
		"short_url", shortURL,
		"request_id", requestID,
	)

	url, err := s.urlRepository.GetByShortURL(ctx, shortURL)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return "", nil
		}
		logger.Errorw("Failed to extract long URL from storage",
			"error", err,
			"short_url", shortURL,
			"request_id", requestID,
		)
		return "", err
	}

	if url == nil {
		logger.Warnw("Short URL not found in storage",
			"short_url", shortURL,
			"request_id", requestID,
		)
		return "", nil
	}

	if url.IsDeleted {
		logger.Warnw("Short URL is deleted from storage",
			"short_url", shortURL,
			"request_id", requestID,
		)
		return "", service.ErrURLWasDeleted
	}

	logger.Infow("Successfully extracted long URL from storage",
		"long_url", url.LongURL,
		"short_url", shortURL,
		"request_id", requestID,
	)

	s.notifyFollowURL(ctx, url.LongURL)
	return url.LongURL, nil
}

func (s *linkExtractorService) notifyFollowURL(ctx context.Context, longURL string) {
	// Если eventBus не инициализирован, пропускаем отправку события
	if s.eventBus == nil {
		return
	}

	var userID string

	logger := middleware.GetLogger(ctx)
	user := middleware.GetJWTUserFromContext(ctx)
	if user == nil {
		userID = "anonymous"
	} else {
		userID = user.ID
	}

	event := audit.Event{
		Timestamp: int(time.Now().Unix()),
		Action:    audit.EventFollow,
		UserID:    userID,
		URL:       longURL,
	}
	go func() {
		if err := s.eventBus.NotifyAll(ctx, event); err != nil {
			logger.Errorw("Failed to send follow URL notification event err=", err)
		}
	}()
}
