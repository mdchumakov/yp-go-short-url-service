//go:generate $HOME/go/bin/mockgen -source=interfaces.go -destination=mock/mock_service.go -package=mock

package service

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/model"
)

type URLShortenerService interface {
	ShortURL(ctx context.Context, longURL string) (string, error)
	ShortURLsByBatch(ctx context.Context, longURLs []map[string]string) ([]map[string]string, error)
}

type URLExtractorService interface {
	ExtractLongURL(ctx context.Context, shortURL string) (string, error)
	ExtractUserURLs(ctx context.Context, userID string) ([]*model.URLsModel, error)
}

type URLDestructorService interface {
	DeleteURL(ctx context.Context, shortURL string) error
	DeleteURLsByBatch(ctx context.Context, shortURLs []string) error
	Stop()
}

type HealthCheckService interface {
	Ping(ctx context.Context) error
}

type DataInitializerService interface {
	Setup(ctx context.Context, fileStoragePath string) error
}

type AuthService interface {
	CreateUser(ctx context.Context, username, password string) (*model.UserModel, error)
	CreateAnonymousUser(ctx context.Context, clientIP, userAgent string, expiresAt *time.Time) (*model.UserModel, error)
	GetOrCreateAnonymousUser(ctx context.Context, clientIP, userAgent string) (*model.UserModel, error)
	GetUserByID(ctx context.Context, userID string) (*model.UserModel, error)
	GetUserByName(ctx context.Context, username string) (*model.UserModel, error)
	GenerateAnonymousName(clientIP, userAgent string) string
}

type JWTService interface {
	GenerateToken(ctx context.Context, username string) (string, error)
	GenerateTokenForUser(ctx context.Context, user *model.UserModel) (string, error)
	ValidateToken(ctx context.Context, token string) (*model.UserModel, error)
	GetUserIDFromToken(ctx context.Context, token string) (string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	IsTokenExpired(ctx context.Context, token string) (bool, error)
	GetTokenExpirationTime(ctx context.Context, token string) (*time.Time, error)
}
