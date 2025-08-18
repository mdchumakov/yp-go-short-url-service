package postgres

import (
	"context"
	"errors"
	"fmt"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userURLsRepository struct {
	pool PoolInterface
}

func NewUserURLsRepository(pool *pgxpool.Pool) repository.UserURLsRepository {
	return &userURLsRepository{pool: pool}
}

func (r *userURLsRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error) {
	query := `
		SELECT u.id, u.short_url, u.long_url, u.created_at, u.updated_at
		FROM urls u
		INNER JOIN user_urls uu ON u.id = uu.url_id
		WHERE uu.user_id = $1
		ORDER BY uu.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
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

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Отложенный rollback (выполнится только если не будет commit)
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				// Логируем ошибку rollback, но возвращаем оригинальную ошибку
				fmt.Printf("rollback failed: %v\n", rollbackErr)
			}
		}
	}()

	// 1. Создаем URL
	urlQuery := `INSERT INTO urls (short_url, long_url) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRow(ctx, urlQuery, url.ShortURL, url.LongURL).Scan(&url.ID)
	if err != nil {
		// Проверяем на дублирование записи
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrURLExists
		}
		return fmt.Errorf("failed to insert url: %w", err)
	}

	// 2. Связываем URL с пользователем
	userURLQuery := `INSERT INTO user_urls (user_id, url_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, userURLQuery, userID, url.ID)
	if err != nil {
		// Проверяем на дублирование записи
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrURLExists
		}
		return fmt.Errorf("failed to link url to user: %w", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Отложенный rollback
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				fmt.Printf("rollback failed: %v\n", rollbackErr)
			}
		}
	}()

	// Подготавливаем batch запросы
	urlQuery := `INSERT INTO urls (short_url, long_url) VALUES ($1, $2) RETURNING id`
	userURLQuery := `INSERT INTO user_urls (user_id, url_id) VALUES ($1, $2)`

	// Выполняем batch операцию
	for _, url := range urls {
		if url == nil {
			continue
		}

		// Создаем URL
		err = tx.QueryRow(ctx, urlQuery, url.ShortURL, url.LongURL).Scan(&url.ID)
		if err != nil {
			// Проверяем на дублирование записи
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
				return repository.ErrURLExists
			}
			return fmt.Errorf("failed to insert url %s: %w", url.ShortURL, err)
		}

		// Связываем с пользователем
		_, err = tx.Exec(ctx, userURLQuery, userID, url.ID)
		if err != nil {
			// Проверяем на дублирование записи
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
				return repository.ErrURLExists
			}
			return fmt.Errorf("failed to link url %s to user: %w", url.ShortURL, err)
		}
	}

	return tx.Commit(ctx)
}
