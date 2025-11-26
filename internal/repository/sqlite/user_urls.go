package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/google/uuid"
)

type userURLsRepository struct {
	db *sql.DB
}

// DeleteURLsWithUser помечает указанные URL как удаленные для конкретного пользователя в SQLite.
// Принимает список коротких URL и идентификатор пользователя, возвращает ошибку, если удаление не удалось.
func (r *userURLsRepository) DeleteURLsWithUser(ctx context.Context, shortURLs []string, userID string) error {
	//TODO implement me
	panic("implement me")
}

// NewUserURLsRepository создает новый репозиторий для работы с URL пользователей в SQLite базе данных.
// Принимает соединение с SQLite и возвращает реализацию интерфейса UserURLsRepository.
func NewUserURLsRepository(db *sql.DB) repository.UserURLsRepository {
	return &userURLsRepository{db: db}
}

// GetByUserID получает все URL, принадлежащие указанному пользователю, из базы данных SQLite.
// Возвращает список моделей URL, отсортированных по дате создания (от новых к старым), или ошибку.
func (r *userURLsRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error) {
	query := `
		SELECT u.id, u.short_url, u.long_url, u.created_at, u.updated_at
		FROM urls u
		INNER JOIN user_urls uu ON u.id = uu.url_id
		WHERE uu.user_id = ?
		ORDER BY uu.created_at DESC
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

// CreateURLWithUser создает новую запись URL и связывает ее с пользователем в базе данных SQLite.
// Принимает модель URL и идентификатор пользователя, возвращает ошибку, если создание не удалось.
func (r *userURLsRepository) CreateURLWithUser(ctx context.Context, url *model.URLsModel, userID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	if url == nil {
		return errors.New("url cannot be nil")
	}

	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Отложенный rollback (выполнится только если не будет commit)
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				// Логируем ошибку rollback, но возвращаем оригинальную ошибку
				fmt.Printf("rollback failed: %v\n", rollbackErr)
			}
		}
	}()

	// 1. Создаем URL
	urlQuery := `INSERT INTO urls (short_url, long_url, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now'))`
	result, err := tx.ExecContext(ctx, urlQuery, url.ShortURL, url.LongURL)
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

	// 2. Связываем URL с пользователем
	userURLQuery := `INSERT INTO user_urls (id, user_id, url_id) VALUES (?, ?, ?)`
	id := uuid.New()
	_, err = tx.ExecContext(ctx, userURLQuery, id.String(), userID, url.ID)
	if err != nil {
		// Проверяем на дублирование записи
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return repository.ErrURLExists
		}
		return fmt.Errorf("failed to link url to user: %w", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateMultipleURLsWithUser создает несколько записей URL и связывает их с пользователем в одной транзакции в SQLite.
// Принимает список моделей URL и идентификатор пользователя, возвращает ошибку, если создание не удалось.
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

	// Подготавливаем batch запросы
	urlQuery := `INSERT INTO urls (short_url, long_url, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now'))`
	userURLQuery := `INSERT INTO user_urls (user_id, url_id) VALUES (?, ?)`

	// Выполняем batch операцию
	for _, url := range urls {
		if url == nil {
			continue
		}

		// 1. Создаем URL
		result, err := tx.ExecContext(ctx, urlQuery, url.ShortURL, url.LongURL)
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

		// 2. Связываем с пользователем
		_, err = tx.ExecContext(ctx, userURLQuery, userID, url.ID)
		if err != nil {
			// Проверяем на дублирование записи
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return repository.ErrURLExists
			}
			return fmt.Errorf("failed to link url %s to user: %w", url.ShortURL, err)
		}
	}

	return tx.Commit()
}
