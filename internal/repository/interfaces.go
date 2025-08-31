//go:generate mockgen -source=interfaces.go -destination=mock/mock_repository.go -package=mock

package repository

import (
	"context"
	"time"
	"yp-go-short-url-service/internal/model"
)

type URLRepository interface {
	URLRepositoryReader
	URLRepositoryWriter
}

type URLRepositoryReader interface {
	Ping(ctx context.Context) error
	GetByLongURL(ctx context.Context, longURL string) (*model.URLsModel, error)
	GetByShortURL(ctx context.Context, shortURL string) (*model.URLsModel, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.URLsModel, error)
	GetTotalCount(ctx context.Context) (int64, error)
}

type URLRepositoryWriter interface {
	Create(ctx context.Context, url *model.URLsModel) error
	CreateBatch(ctx context.Context, urls []*model.URLsModel) error
}

type UserRepository interface {
	UserRepositoryReader
	UserRepositoryCreator
}

type UserRepositoryCreator interface {
	CreateUser(ctx context.Context, username, password string, expiresAt *time.Time) (*model.UserModel, error)
}

type UserRepositoryReader interface {
	GetUserByID(ctx context.Context, userID string) (*model.UserModel, error)
	GetUserByName(ctx context.Context, username string) (*model.UserModel, error)
}

type UserURLsRepository interface {
	UserURLsRepositoryReader
	UserURLsRepositoryWriter
}

type UserURLsRepositoryReader interface {
	GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error)
}

type UserURLsRepositoryWriter interface {
	CreateURLWithUser(ctx context.Context, url *model.URLsModel, userID string) error
	CreateMultipleURLsWithUser(ctx context.Context, urls []*model.URLsModel, userID string) error
	DeleteURLsWithUser(ctx context.Context, shortURLs []string, userID string) error
}
