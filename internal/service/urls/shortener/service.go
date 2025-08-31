package shortener

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func NewURLShortenerService(
	urlRepository repository.URLRepository,
	userURLsRepository repository.UserURLsRepository,
) service.URLShortenerService {
	return &urlShortenerService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
	}
}

type urlShortenerService struct {
	urlRepository      repository.URLRepository
	userURLsRepository repository.UserURLsRepository
}

func (s *urlShortenerService) ShortURLsByBatch(ctx context.Context, longURLs []map[string]string) ([]map[string]string, error) {
	logger := middleware.GetLogger(ctx)

	urlsForCreation, err := s.processURLsForBatch(ctx, longURLs)
	if err != nil {
		return nil, err
	}

	if err := s.saveURLsToStorage(ctx, urlsForCreation); err != nil {
		return nil, err
	}

	logger.Infow("URL records created successfully in database")
	return longURLs, nil
}

// processURLsForBatch обрабатывает массив URL'ов для пакетного создания
func (s *urlShortenerService) processURLsForBatch(ctx context.Context, longURLs []map[string]string) ([]*model.URLsModel, error) {
	var urlsForCreation []*model.URLsModel

	for _, longURLItem := range longURLs {
		longURL := longURLItem["original_url"]

		processedURL, err := s.processSingleURL(ctx, longURLItem, longURL)
		if err != nil {
			return nil, err
		}

		if processedURL != nil {
			urlsForCreation = append(urlsForCreation, processedURL)
		}
	}

	return urlsForCreation, nil
}

// processSingleURL обрабатывает один URL из пакета
func (s *urlShortenerService) processSingleURL(ctx context.Context, longURLItem map[string]string, longURL string) (*model.URLsModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	logger.Infow("Starting URL shortening process",
		"long_url", longURL,
		"request_id", requestID,
	)

	// Проверяем, существует ли уже короткий URL
	shortURLFromStorage, err := s.extractShortURLIfExists(ctx, longURL)
	if err != nil {
		logger.Errorw("Failed to extract short URL from storage",
			"error", err,
			"long_url", longURL,
			"request_id", requestID,
		)
		return nil, err
	}

	// Если URL уже существует, используем его
	if shortURLFromStorage != nil {
		return s.handleExistingURL(ctx, longURLItem, longURL, *shortURLFromStorage)
	}

	// Создаем новый короткий URL
	return s.createNewShortURL(ctx, longURLItem, longURL)
}

// handleExistingURL обрабатывает случай, когда короткий URL уже существует
func (s *urlShortenerService) handleExistingURL(
	ctx context.Context,
	longURLItem map[string]string,
	longURL,
	shortURL string,
) (*model.URLsModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	logger.Infow("Short URL already exists in storage",
		"short_url", shortURL,
		"long_url", longURL,
		"request_id", requestID,
	)

	longURLItem["short_url"] = shortURL
	return nil, nil // Не создаем новую запись
}

// createNewShortURL создает новый короткий URL
func (s *urlShortenerService) createNewShortURL(ctx context.Context, longURLItem map[string]string, longURL string) (*model.URLsModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	shortURL := shortenURLBase62(longURL)
	logger.Debugw("Generated short URL",
		"short_url", shortURL,
		"request_id", requestID,
	)

	// Добавляем short_url в результат
	longURLItem["short_url"] = shortURL

	newURL := model.URLsModel{
		ShortURL:  shortURL,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &newURL, nil
}

// saveURLsToStorage сохраняет URL'ы в базу данных
func (s *urlShortenerService) saveURLsToStorage(ctx context.Context, urlsForCreation []*model.URLsModel) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)
	user := middleware.GetJWTUserFromContext(ctx)

	if user != nil {
		logger.Infow("Associating URLs with user",
			"user", user,
			"request_id", requestID,
		)
		if err := s.userURLsRepository.CreateMultipleURLsWithUser(ctx, urlsForCreation, user.ID); err != nil {
			logger.Errorw("Failed to associate URLs with user",
				"error", err,
			)
			return err
		}
	} else {
		logger.Warnw("JWT user is nil, skipping user association",
			"request_id", requestID,
		)
		if err := s.urlRepository.CreateBatch(ctx, urlsForCreation); err != nil {
			logger.Errorw("Failed to create URL records in database",
				"error", err,
				"request_id", requestID,
			)
			return err
		}
	}

	return nil
}

func (s *urlShortenerService) ShortURL(ctx context.Context, longURL string) (string, error) {
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
		return *shortURLFromStorage, service.ErrURLAlreadyExists
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

func (s *urlShortenerService) extractShortURLIfExists(ctx context.Context, longURL string) (*string, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	urlResponse, err := s.urlRepository.GetByLongURL(ctx, longURL)
	if err != nil {
		logger.Debugw("Failed to extract short URL from storage", "error", err, "request_id", requestID)
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

func (s *urlShortenerService) saveShortURLToStorage(ctx context.Context, url *model.URLsModel) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)
	user := middleware.GetJWTUserFromContext(ctx)

	if user != nil {
		logger.Debugw("Associating user with URL", "user", user, "request_id", requestID)
		err := s.userURLsRepository.CreateURLWithUser(ctx, url, user.ID)
		if err != nil {
			logger.Errorw("Failed to associate user with URL",
				"error", err,
			)
		}
	} else {
		logger.Warnw("JWT user is nil, skipping user association",
			"short_url", url.ShortURL,
			"request_id", requestID,
		)
		err := s.urlRepository.Create(ctx, url)
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
	}
	return nil
}
