//go:generate $HOME/go/bin/mockgen -source=interfaces.go -destination=mock/mock_repository.go -package=mock

package repository

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/model"
)

// URLRepository определяет полный интерфейс для работы с URL в базе данных.
// Объединяет интерфейсы для чтения и записи URL.
type URLRepository interface {
	URLRepositoryReader
	URLRepositoryWriter
}

// URLRepositoryReader определяет интерфейс для чтения URL из базы данных.
// Предоставляет методы для получения URL по различным критериям и проверки соединения с БД.
type URLRepositoryReader interface {
	Ping(ctx context.Context) error
	GetByLongURL(ctx context.Context, longURL string) (*model.URLsModel, error)
	GetByShortURL(ctx context.Context, shortURL string) (*model.URLsModel, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.URLsModel, error)
	GetTotalCount(ctx context.Context) (int64, error)
}

// URLRepositoryWriter определяет интерфейс для записи URL в базу данных.
// Предоставляет методы для создания одного или нескольких URL.
type URLRepositoryWriter interface {
	Create(ctx context.Context, url *model.URLsModel) error
	CreateBatch(ctx context.Context, urls []*model.URLsModel) error
}

// UserRepository определяет полный интерфейс для работы с пользователями в базе данных.
// Объединяет интерфейсы для чтения и создания пользователей.
type UserRepository interface {
	UserRepositoryReader
	UserRepositoryCreator
}

// UserRepositoryCreator определяет интерфейс для создания пользователей в базе данных.
// Предоставляет метод для создания нового пользователя с указанными параметрами.
type UserRepositoryCreator interface {
	CreateUser(ctx context.Context, username, password string, expiresAt *time.Time) (*model.UserModel, error)
}

// UserRepositoryReader определяет интерфейс для чтения пользователей из базы данных.
// Предоставляет методы для получения пользователя по ID или имени.
type UserRepositoryReader interface {
	GetUserByID(ctx context.Context, userID string) (*model.UserModel, error)
	GetUserByName(ctx context.Context, username string) (*model.UserModel, error)
}

// UserURLsRepository определяет полный интерфейс для работы с URL пользователей в базе данных.
// Объединяет интерфейсы для чтения и записи связей между пользователями и URL.
type UserURLsRepository interface {
	UserURLsRepositoryReader
	UserURLsRepositoryWriter
}

// UserURLsRepositoryReader определяет интерфейс для чтения URL пользователей из базы данных.
// Предоставляет метод для получения всех URL, принадлежащих конкретному пользователю.
type UserURLsRepositoryReader interface {
	GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error)
}

// UserURLsRepositoryWriter определяет интерфейс для записи связей между пользователями и URL в базу данных.
// Предоставляет методы для создания связей и удаления URL пользователя.
type UserURLsRepositoryWriter interface {
	CreateURLWithUser(ctx context.Context, url *model.URLsModel, userID string) error
	CreateMultipleURLsWithUser(ctx context.Context, urls []*model.URLsModel, userID string) error
	DeleteURLsWithUser(ctx context.Context, shortURLs []string, userID string) error
}
