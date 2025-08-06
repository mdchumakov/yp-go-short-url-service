package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

// DBInterface определяет интерфейс для работы с базой данных
type DBInterface interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PingContext(ctx context.Context) error
}

type urlsRepository struct {
	db DBInterface
}

func NewURLsRepository(db *sql.DB) repository.URLRepository {
	return &urlsRepository{db: db}
}

func (r *urlsRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *urlsRepository) GetByLongURL(ctx context.Context, longURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `SELECT id, short_url, long_url, created_at, updated_at FROM urls WHERE long_url = ?`

	err := r.db.QueryRowContext(ctx, query, longURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrURLNotFound
		}
		return nil, err
	}

	return &urls, nil
}

func (r *urlsRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `SELECT id, short_url, long_url, created_at, updated_at FROM urls WHERE short_url = ?`

	err := r.db.QueryRowContext(ctx, query, shortURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrURLNotFound
		}
		return nil, err
	}

	return &urls, nil
}

func (r *urlsRepository) Create(ctx context.Context, url *model.URLsModel) error {
	if url == nil {
		return errors.New("url cannot be nil")
	}

	query := `INSERT INTO urls (short_url, long_url, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now'))`

	_, err := r.db.ExecContext(ctx, query, url.ShortURL, url.LongURL)
	if err != nil {
		return err
	}

	return nil
}
