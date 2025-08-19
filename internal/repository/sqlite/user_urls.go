package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
)

type userURLsRepository struct {
	db *sql.DB
}

func NewUserURLsRepository(db *sql.DB) repository.UserURLsRepository {
	return &userURLsRepository{db: db}
}

func (r *userURLsRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error) {
	query := `
		SELECT id, short_url, long_url, created_at, updated_at
		FROM urls
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *userURLsRepository) CreateURLWithUser(ctx context.Context, url *model.URLsModel, userID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	if url == nil {
		return errors.New("url cannot be nil")
	}

	// Создаем URL с user_id
	urlQuery := `INSERT INTO urls (short_url, long_url, user_id, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`
	result, err := r.db.ExecContext(ctx, urlQuery, url.ShortURL, url.LongURL, userID)
	if err != nil {
		// Проверяем на дублирование записи в SQLite
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return repository.ErrURLExists
		}
		return fmt.Errorf("failed to insert url: %w", err)
	}

	// Получаем ID созданного URL
	urlID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	url.ID = uint(urlID)

	return nil
}

func (r *userURLsRepository) CreateMultipleURLsWithUser(ctx context.Context, urls []*model.URLsModel, userID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	if len(urls) == 0 {
		return nil
	}

	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Отложенный rollback
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				fmt.Printf("rollback failed: %v\n", rollbackErr)
			}
		}
	}()

	// Подготавливаем batch запрос
	urlQuery := `INSERT INTO urls (short_url, long_url, user_id, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`

	// Выполняем batch операцию
	for _, url := range urls {
		if url == nil {
			continue
		}

		// Создаем URL с user_id
		result, err := tx.ExecContext(ctx, urlQuery, url.ShortURL, url.LongURL, userID)
		if err != nil {
			// Проверяем на дублирование записи в SQLite
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return repository.ErrURLExists
			}
			return fmt.Errorf("failed to insert url %s: %w", url.ShortURL, err)
		}

		// Получаем ID созданного URL
		urlID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id for url %s: %w", url.ShortURL, err)
		}
		url.ID = uint(urlID)
	}

	return tx.Commit()
}
