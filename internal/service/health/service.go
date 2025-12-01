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

// NewHealthCheckService создает новый сервис проверки здоровья приложения.
// Принимает репозиторий для чтения URL и возвращает реализацию интерфейса HealthCheckService.
func NewHealthCheckService(repo repository.URLRepositoryReader) service.HealthCheckService {
	return &healthCheckService{
		repository: repo,
	}
}

// Ping проверяет доступность базы данных, выполняя простой запрос.
// Возвращает ошибку, если соединение с базой данных недоступно.
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
