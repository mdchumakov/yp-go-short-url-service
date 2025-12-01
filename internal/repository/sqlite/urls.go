package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

// DBInterface определяет интерфейс для работы с базой данных SQLite.
// Используется для абстракции работы с базой данных и упрощения тестирования.
type DBInterface interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PingContext(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type urlsRepository struct {
	db DBInterface
}

// NewURLsRepository создает новый репозиторий для работы с URL в SQLite базе данных.
// Принимает соединение с SQLite и возвращает реализацию интерфейса URLRepository.
func NewURLsRepository(db *sql.DB) repository.URLRepository {
	return &urlsRepository{db: db}
}

// Ping проверяет доступность базы данных SQLite.
// Возвращает ошибку, если соединение с базой данных недоступно.
func (r *urlsRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// GetByLongURL получает URL из базы данных SQLite по длинному URL.
// Возвращает модель URL или ошибку, если URL не найден.
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrURLNotFound
		}
		return nil, err
	}

	return &urls, nil
}

// GetByShortURL получает URL из базы данных SQLite по короткому идентификатору.
// Возвращает модель URL или ошибку, если URL не найден.
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrURLNotFound
		}
		return nil, err
	}

	return &urls, nil
}

// Create создает новую запись URL в базе данных SQLite.
// Принимает модель URL и возвращает ошибку, если создание не удалось.
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

// CreateBatch создает несколько записей URL в базе данных SQLite в одной транзакции.
// Принимает список моделей URL и возвращает ошибку, если создание не удалось.
func (r *urlsRepository) CreateBatch(ctx context.Context, urls []*model.URLsModel) error {
	if len(urls) == 0 {
		return nil
	}

	// Используем транзакцию для атомарности операции
	tx, err := r.db.(*sql.DB).BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	// Подготавливаем batch insert запрос
	query := `INSERT OR IGNORE INTO urls (id, short_url, long_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			return
		}
	}(stmt)

	// Выполняем batch вставку
	for _, url := range urls {
		if url == nil {
			continue
		}

		_, err := stmt.ExecContext(ctx, url.ID, url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt)
		if err != nil {
			return err
		}
	}

	// Подтверждаем транзакцию
	return tx.Commit()
}

// GetAll получает список URL из базы данных SQLite с пагинацией.
// Принимает лимит и смещение для пагинации, возвращает список моделей URL или ошибку.
func (r *urlsRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.URLsModel, error) {
	query := `SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

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

// GetTotalCount получает общее количество URL в базе данных SQLite.
// Возвращает количество записей или ошибку, если запрос не удался.
func (r *urlsRepository) GetTotalCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM urls`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
