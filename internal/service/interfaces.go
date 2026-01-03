//go:generate $HOME/go/bin/mockgen -source=interfaces.go -destination=mock/mock_service.go -package=mock

package service

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/model"
)

// URLShortenerService определяет интерфейс для сервиса сокращения URL.
// Предоставляет методы для создания коротких ссылок из длинных URL.
type URLShortenerService interface {
	ShortURL(ctx context.Context, longURL string) (string, error)
	ShortURLsByBatch(ctx context.Context, longURLs []map[string]string) ([]map[string]string, error)
}

// URLExtractorService определяет интерфейс для сервиса извлечения URL.
// Предоставляет методы для получения длинных URL по коротким и для получения всех URL пользователя.
type URLExtractorService interface {
	ExtractLongURL(ctx context.Context, shortURL string) (string, error)
	ExtractUserURLs(ctx context.Context, userID string) ([]*model.URLsModel, error)
}

// URLDestructorService определяет интерфейс для сервиса удаления URL.
// Предоставляет методы для асинхронного удаления URL пользователя.
type URLDestructorService interface {
	DeleteURL(ctx context.Context, shortURL string) error
	DeleteURLsByBatch(ctx context.Context, shortURLs []string) error
	Stop()
}

// HealthCheckService определяет интерфейс для сервиса проверки здоровья приложения.
// Используется для проверки доступности базы данных.
type HealthCheckService interface {
	Ping(ctx context.Context) error
}

// DataInitializerService определяет интерфейс для сервиса инициализации данных.
// Используется для первоначальной загрузки данных из файлового хранилища в базу данных.
type DataInitializerService interface {
	Setup(ctx context.Context, fileStoragePath string) error
}

// AuthService определяет интерфейс для сервиса аутентификации.
// Предоставляет методы для создания пользователей, получения пользователей и генерации анонимных имен.
type AuthService interface {
	CreateUser(ctx context.Context, username, password string) (*model.UserModel, error)
	CreateAnonymousUser(ctx context.Context, clientIP, userAgent string, expiresAt *time.Time) (*model.UserModel, error)
	GetOrCreateAnonymousUser(ctx context.Context, clientIP, userAgent string) (*model.UserModel, error)
	GetUserByID(ctx context.Context, userID string) (*model.UserModel, error)
	GetUserByName(ctx context.Context, username string) (*model.UserModel, error)
	GenerateAnonymousName(clientIP, userAgent string) string
}

// JWTService определяет интерфейс для работы с JWT токенами.
// Предоставляет методы для генерации, валидации, обновления токенов и проверки их срока действия.
type JWTService interface {
	GenerateToken(ctx context.Context, username string) (string, error)
	GenerateTokenForUser(ctx context.Context, user *model.UserModel) (string, error)
	ValidateToken(ctx context.Context, token string) (*model.UserModel, error)
	GetUserIDFromToken(ctx context.Context, token string) (string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	IsTokenExpired(ctx context.Context, token string) (bool, error)
	GetTokenExpirationTime(ctx context.Context, token string) (*time.Time, error)
}

// StatsService определяет интерфейс для сервиса статистики.
// Предоставляет методы для получения общей статистики по URL и пользователям.
type StatsService interface {
	GetTotalURLsCount(ctx context.Context) (int64, error)
	GetTotalUsersCount(ctx context.Context) (int64, error)
}
