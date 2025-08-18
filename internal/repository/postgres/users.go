package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &usersRepository{pool: pool}
}

// CreateUser создает нового пользователя
func (r *usersRepository) CreateUser(ctx context.Context, username, password string, expiresAt *time.Time) (*model.UserModel, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	query := `
		INSERT INTO users (name, password, is_anonymous, expires_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, name, password, is_anonymous, expires_at, created_at, updated_at
	`

	var user model.UserModel
	isAnonymous := password == ""

	err := r.pool.QueryRow(ctx, query, username, password, isAnonymous, expiresAt, time.Now(), time.Now()).Scan(
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
		INSERT INTO users (name, password, is_anonymous, expires_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, name, password, is_anonymous, expires_at, created_at, updated_at
	`

	var user model.UserModel
	err := r.pool.QueryRow(ctx, query, anonymousName, nil, true, time.Now(), time.Now()).Scan(
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
	query := `SELECT id, name, password, is_anonymous, created_at, updated_at FROM users WHERE id = $1`

	var user model.UserModel
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByName получает пользователя по имени
func (r *usersRepository) GetUserByName(ctx context.Context, username string) (*model.UserModel, error) {
	query := `SELECT * FROM users WHERE name = $1`

	var user model.UserModel
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.ExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CleanupExpiredAnonymousUsers удаляет устаревших анонимных пользователей (если используется TTL)
func (r *usersRepository) CleanupExpiredAnonymousUsers(ctx context.Context) error {
	query := `DELETE FROM users WHERE is_anonymous = true AND expires_at < NOW()`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired anonymous users: %w", err)
	}

	return nil
}
