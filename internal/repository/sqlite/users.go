package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/google/uuid"
)

type usersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) repository.UserRepository {
	return &usersRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *usersRepository) CreateUser(ctx context.Context, username, password string, expiresAt *time.Time) (*model.UserModel, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	query := `
		INSERT INTO users (name, password, is_anonymous, expires_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?) 
		RETURNING id, name, password, is_anonymous, expires_at, created_at, updated_at
	`

	var user model.UserModel
	isAnonymous := password == ""

	err := r.db.QueryRowContext(ctx, query, username, password, isAnonymous).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.ExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// CreateAnonymousUser создает анонимного пользователя с уникальным именем
func (r *usersRepository) CreateAnonymousUser(ctx context.Context) (*model.UserModel, error) {
	// Генерируем уникальное имя для анонимного пользователя
	anonymousName := fmt.Sprintf("anonymous_%s", uuid.New().String())

	query := `
		INSERT INTO users (name, password, is_anonymous, created_at, updated_at) 
		VALUES (?, ?, ?, datetime('now'), datetime('now')) 
		RETURNING id, name, password, is_anonymous, created_at, updated_at
	`

	var user model.UserModel
	err := r.db.QueryRowContext(ctx, query, anonymousName, nil, true).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create anonymous user: %w", err)
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *usersRepository) GetUserByID(ctx context.Context, userID string) (*model.UserModel, error) {
	query := `SELECT id, name, password, is_anonymous, created_at, updated_at FROM users WHERE id = ?`

	var user model.UserModel
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByName получает пользователя по имени
func (r *usersRepository) GetUserByName(ctx context.Context, username string) (*model.UserModel, error) {
	query := `SELECT id, name, password, is_anonymous, created_at, updated_at FROM users WHERE name = ?`

	var user model.UserModel
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CleanupExpiredAnonymousUsers удаляет устаревших анонимных пользователей (если используется TTL)
func (r *usersRepository) CleanupExpiredAnonymousUsers(ctx context.Context) error {
	query := `DELETE FROM users WHERE is_anonymous = 1 AND expires_at < datetime('now')`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired anonymous users: %w", err)
	}

	return nil
}
