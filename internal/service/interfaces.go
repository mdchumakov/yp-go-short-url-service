//go:generate mockgen -source=interfaces.go -destination=mock/mock_service.go -package=mock

package service

import (
	"context"
)

type LinkShortenerService interface {
	ShortURL(ctx context.Context, longURL string) (string, error)
}

type LinkExtractorService interface {
	ExtractLongURL(ctx context.Context, shortURL string) (string, error)
}

type HealthCheckService interface {
	Ping(ctx context.Context) error
}

type DataInitializerService interface {
	Setup(ctx context.Context, fileStoragePath string) error
}
