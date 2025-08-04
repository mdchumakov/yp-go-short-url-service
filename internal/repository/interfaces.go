//go:generate mockgen -source=interfaces.go -destination=mock/mock_repository.go -package=mock

package repository

import (
	"context"
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
}

type URLRepositoryWriter interface {
	Create(ctx context.Context, url *model.URLsModel) error
}
