package postgres

import (
	"context"
	"errors"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolInterface определяет интерфейс для пула соединений
type PoolInterface interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type urlsRepository struct {
	pool PoolInterface
}

func NewURLsRepository(pool *pgxpool.Pool) repository.URLRepository {
	return &urlsRepository{pool: pool}
}

func (r *urlsRepository) Ping(ctx context.Context) error {
	query := `SELECT 1`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (r *urlsRepository) GetByLongURL(ctx context.Context, longURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `SELECT * FROM urls WHERE long_url = $1`

	err := r.pool.QueryRow(ctx, query, longURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &urls, nil
}

func (r *urlsRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `SELECT * FROM urls WHERE short_url = $1`

	err := r.pool.QueryRow(ctx, query, shortURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &urls, nil
}

func (r *urlsRepository) Create(ctx context.Context, url *model.URLsModel) error {
	if url == nil {
		return errors.New("url cannot be nil")
	}

	query := `INSERT INTO urls (short_url, long_url) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, url.ShortURL, url.LongURL)
	if err != nil {
		return err
	}

	return nil
}
