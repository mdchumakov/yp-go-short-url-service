package destructor

import (
	"context"
	"errors"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func NewURLDestructorService(
	urlRepository repository.URLRepository,
	userURLsRepository repository.UserURLsRepository,
) service.URLDestructorService {
	return &urlDestructorService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
	}
}

type urlDestructorService struct {
	urlRepository      repository.URLRepository
	userURLsRepository repository.UserURLsRepository
}

func (s *urlDestructorService) DeleteURL(ctx context.Context, shortURL string) error {
	//TODO implement me
	panic("implement me")
}

func (s *urlDestructorService) DeleteURLsByBatch(ctx context.Context, shortURLs []string) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	user := middleware.GetJWTUserFromContext(ctx)
	if user == nil {
		return errors.New("user is not authenticated")
	}

	logger.Debugw("Starting URL deletion process",
		"request_id", requestID,
	)

	err := s.userURLsRepository.DeleteURLsWithUser(ctx, shortURLs, user.ID)
	if err != nil {
		logger.Errorw("Failed to delete URLs from storage",
			"error", err,
			"request_id", requestID,
		)
		return err
	}

	return nil
}
