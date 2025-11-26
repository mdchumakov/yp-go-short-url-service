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

// PoolInterface определяет интерфейс для пула соединений PostgreSQL.
// Используется для абстракции работы с базой данных и упрощения тестирования.
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

// NewURLsRepository создает новый репозиторий для работы с URL в PostgreSQL базе данных.
// Принимает пул соединений PostgreSQL и возвращает реализацию интерфейса URLRepository.
func NewURLsRepository(pool *pgxpool.Pool) repository.URLRepository {
	return &urlsRepository{pool: pool}
}

// Ping проверяет доступность базы данных PostgreSQL, выполняя простой запрос SELECT 1.
// Возвращает ошибку, если соединение с базой данных недоступно.
func (r *urlsRepository) Ping(ctx context.Context) error {
	query := `SELECT 1`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

// GetByLongURL получает URL из базы данных по длинному URL.
// Возвращает модель URL или ошибку, если URL не найден или был удален.
func (r *urlsRepository) GetByLongURL(ctx context.Context, longURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `
		SELECT id, short_url, long_url, is_deleted, created_at, updated_at 
		FROM urls 
		WHERE long_url = $1 AND is_deleted = false
		`

	err := r.pool.QueryRow(ctx, query, longURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.IsDeleted,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &urls, nil
}

// GetByShortURL получает URL из базы данных по короткому идентификатору.
// Возвращает модель URL или ошибку, если URL не найден.
func (r *urlsRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URLsModel, error) {
	var urls model.URLsModel

	query := `SELECT id, short_url, long_url, is_deleted, created_at, updated_at FROM urls WHERE short_url = $1`

	err := r.pool.QueryRow(ctx, query, shortURL).Scan(
		&urls.ID,
		&urls.ShortURL,
		&urls.LongURL,
		&urls.IsDeleted,
		&urls.CreatedAt,
		&urls.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &urls, nil
}

// Create создает новую запись URL в базе данных.
// Принимает модель URL и возвращает ошибку, если создание не удалось.
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

// CreateBatch создает несколько записей URL в базе данных в одной транзакции.
// Принимает список моделей URL и возвращает ошибку, если создание не удалось.
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

// GetAll получает список URL из базы данных с пагинацией.
// Принимает лимит и смещение для пагинации, возвращает список моделей URL или ошибку.
func (r *urlsRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.URLsModel, error) {
	query := `
		SELECT id, short_url, long_url, is_deleted, created_at, updated_at
		FROM urls 
		WHERE is_deleted = false
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

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
			&url.IsDeleted,
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

// GetTotalCount получает общее количество URL в базе данных.
// Возвращает количество записей или ошибку, если запрос не удался.
func (r *urlsRepository) GetTotalCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM urls WHERE is_deleted = false`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *urlsRepository) SoftDeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	// Используем транзакцию для атомарности операции
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
		}
	}()

	// Подготавливаем batch update запрос с проверкой владельца
	query := `
		UPDATE urls 
		SET is_deleted = true, updated_at = NOW() 
		WHERE short_url = ANY($1) 
		AND id IN (
			SELECT uu.url_id 
			FROM user_urls uu 
			WHERE uu.user_id = $2
		)
	`

	_, err = tx.Exec(ctx, query, shortURLs, userID)
	if err != nil {
		return err
	}

	// Подтверждаем транзакцию
	return tx.Commit(ctx)
}
