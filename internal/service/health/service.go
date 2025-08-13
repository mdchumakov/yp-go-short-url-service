package health

import (
	"context"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

type healthCheckService struct {
	repository repository.URLRepositoryReader
}

// NewHealthCheckService создает новый сервис проверки здоровья
func NewHealthCheckService(repo repository.URLRepositoryReader) service.HealthCheckService {
	return &healthCheckService{
		repository: repo,
	}
}

func (s *healthCheckService) Ping(ctx context.Context) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	err := s.repository.Ping(ctx)
	if err != nil {
		logger.Errorw("Database connection check failed", "request_id", requestID, "error", err)
		return err
	}
	logger.Infow("Database connection is OK", "request_id", requestID)
	return nil
}
