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
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
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

func (r *urlsRepository) CreateBatch(ctx context.Context, urls []*model.URLsModel) error {
	if len(urls) == 0 {
		return nil
	}

	// Используем транзакцию для атомарности операции
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Подготавливаем batch insert запрос
	query := `INSERT INTO urls (short_url, long_url, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT (short_url) DO NOTHING`

	// Выполняем вставку каждого URL в транзакции
	for _, url := range urls {
		if url == nil {
			continue
		}
		_, err := tx.Exec(ctx, query, url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt)
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return err
			}
			return err
		}
	}

	// Подтверждаем транзакцию
	return tx.Commit(ctx)
}

func (r *urlsRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.URLsModel, error) {
	query := `SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*model.URLsModel
	for rows.Next() {
		var url model.URLsModel
		err := rows.Scan(
			&url.ID,
			&url.ShortURL,
			&url.LongURL,
			&url.CreatedAt,
			&url.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *urlsRepository) GetTotalCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM urls`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
