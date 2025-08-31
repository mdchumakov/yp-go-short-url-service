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
func (r *usersRepository) CreateUser(
	ctx context.Context,
	username, password string,
	expiresAt *time.Time,
) (*model.UserModel, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	query := `
		INSERT INTO users (id, name, password, is_anonymous, expires_at) 
		VALUES (?, ?, ?, ?, ?) 
		RETURNING id, name, password, is_anonymous, expires_at, created_at, updated_at
	`

	var user model.UserModel
	isAnonymous := password == ""

	id := uuid.New()
	err := r.db.QueryRowContext(ctx, query, id.String(), username, password, isAnonymous, expiresAt).Scan(
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

// GetUserByID получает пользователя по ID
func (r *usersRepository) GetUserByID(ctx context.Context, userID string) (*model.UserModel, error) {
	query := `SELECT id, name, password, is_anonymous, expires_at, created_at, updated_at FROM users WHERE id = ?`

	var user model.UserModel
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.ExpiresAt,
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
	query := `SELECT id, name, password, is_anonymous, expires_at, created_at, updated_at FROM users WHERE name = ?`

	var user model.UserModel
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.IsAnonymous,
		&user.ExpiresAt,
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
