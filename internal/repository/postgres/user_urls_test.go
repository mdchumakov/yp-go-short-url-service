package postgres

import (
	"context"
	"testing"
	"time"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserURLsMockPool(t *testing.T) (pgxmock.PgxPoolIface, *userURLsRepository) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	repo := &userURLsRepository{pool: mock}
	return mock, repo
}

func TestUserURLsRepository_GetByUserID(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"

	expectedURLs := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example.com/1",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			ShortURL:  "def456",
			LongURL:   "https://example.com/2",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "is_deleted", "created_at", "updated_at"})
	for _, url := range expectedURLs {
		rows.AddRow(url.ID, url.ShortURL, url.LongURL, url.IsDeleted, url.CreatedAt, url.UpdatedAt)
	}

	mock.ExpectQuery("SELECT u\\.id, u\\.short_url, u\\.long_url, u\\.is_deleted, u\\.created_at, u\\.updated_at FROM urls u INNER JOIN user_urls uu ON u\\.id = uu\\.url_id WHERE uu\\.user_id = \\$1 ORDER BY uu\\.created_at DESC").
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedURLs[0].ID, result[0].ID)
	assert.Equal(t, expectedURLs[0].ShortURL, result[0].ShortURL)
	assert.Equal(t, expectedURLs[0].LongURL, result[0].LongURL)
	assert.Equal(t, expectedURLs[1].ID, result[1].ID)
	assert.Equal(t, expectedURLs[1].ShortURL, result[1].ShortURL)
	assert.Equal(t, expectedURLs[1].LongURL, result[1].LongURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_GetByUserID_EmptyResult(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "is_deleted", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT u\\.id, u\\.short_url, u\\.long_url, u\\.is_deleted, u\\.created_at, u\\.updated_at FROM urls u INNER JOIN user_urls uu ON u\\.id = uu\\.url_id WHERE uu\\.user_id = \\$1 ORDER BY uu\\.created_at DESC").
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_GetByUserID_DatabaseError(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	expectedErr := repository.ErrURLNotFound

	mock.ExpectQuery("SELECT u\\.id, u\\.short_url, u\\.long_url, u\\.is_deleted, u\\.created_at, u\\.updated_at FROM urls u INNER JOIN user_urls uu ON u\\.id = uu\\.url_id WHERE uu\\.user_id = \\$1 ORDER BY uu\\.created_at DESC").
		WithArgs(userID).
		WillReturnError(expectedErr)

	result, err := repo.GetByUserID(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateURLWithUser_Success(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем создание URL
	mock.ExpectQuery("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uint(1)))

	// Ожидаем связывание с пользователем
	mock.ExpectExec("INSERT INTO user_urls \\(user_id, url_id\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(userID, uint(1)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Ожидаем подтверждение транзакции
	mock.ExpectCommit()

	err := repo.CreateURLWithUser(ctx, url, userID)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), url.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateURLWithUser_DuplicateURL(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем ошибку дублирования при создании URL
	pgErr := &pgconn.PgError{
		Code: "23505", // unique_violation
	}

	mock.ExpectQuery("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnError(pgErr)

	// Ожидаем откат транзакции
	mock.ExpectRollback()

	err := repo.CreateURLWithUser(ctx, url, userID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrURLExists, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateURLWithUser_DuplicateUserURL(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем создание URL
	mock.ExpectQuery("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uint(1)))

	// Ожидаем ошибку дублирования при связывании с пользователем
	pgErr := &pgconn.PgError{
		Code: "23505", // unique_violation
	}

	mock.ExpectExec("INSERT INTO user_urls \\(user_id, url_id\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(userID, uint(1)).
		WillReturnError(pgErr)

	// Ожидаем откат транзакции
	mock.ExpectRollback()

	err := repo.CreateURLWithUser(ctx, url, userID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrURLExists, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateURLWithUser_Validation(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	// Тестируем валидацию пустого userID
	err := repo.CreateURLWithUser(ctx, &model.URLsModel{ID: 1}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")

	// Тестируем валидацию nil URL
	err = repo.CreateURLWithUser(ctx, nil, "test-user-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url cannot be nil")
}

func TestUserURLsRepository_CreateMultipleURLsWithUser_Success(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	urls := []*model.URLsModel{
		{ShortURL: "batch1", LongURL: "https://example.com/batch1"},
		{ShortURL: "batch2", LongURL: "https://example.com/batch2"},
		{ShortURL: "batch3", LongURL: "https://example.com/batch3"},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем создание каждого URL
	for i, url := range urls {
		mock.ExpectQuery("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
			WithArgs(url.ShortURL, url.LongURL).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))

		// Ожидаем связывание с пользователем
		mock.ExpectExec("INSERT INTO user_urls \\(user_id, url_id\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs(userID, uint(i+1)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	// Ожидаем подтверждение транзакции
	mock.ExpectCommit()

	err := repo.CreateMultipleURLsWithUser(ctx, urls, userID)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), urls[0].ID)
	assert.Equal(t, uint(2), urls[1].ID)
	assert.Equal(t, uint(3), urls[2].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateMultipleURLsWithUser_EmptySlice(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"

	err := repo.CreateMultipleURLsWithUser(ctx, []*model.URLsModel{}, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateMultipleURLsWithUser_WithNilURLs(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	urls := []*model.URLsModel{
		{ShortURL: "batch1", LongURL: "https://example.com/batch1"},
		nil, // nil URL
		{ShortURL: "batch3", LongURL: "https://example.com/batch3"},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем создание только не-nil URL
	validURLs := []*model.URLsModel{urls[0], urls[2]}
	for i, url := range validURLs {
		mock.ExpectQuery("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
			WithArgs(url.ShortURL, url.LongURL).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))

		// Ожидаем связывание с пользователем
		mock.ExpectExec("INSERT INTO user_urls \\(user_id, url_id\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs(userID, uint(i+1)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	// Ожидаем подтверждение транзакции
	mock.ExpectCommit()

	err := repo.CreateMultipleURLsWithUser(ctx, urls, userID)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), urls[0].ID)
	assert.Equal(t, uint(2), urls[2].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserURLsRepository_CreateMultipleURLsWithUser_Validation(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	urls := []*model.URLsModel{
		{ShortURL: "batch1", LongURL: "https://example.com/batch1"},
	}

	// Тестируем валидацию пустого userID
	err := repo.CreateMultipleURLsWithUser(ctx, urls, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")
}

func TestUserURLsRepository_CreateMultipleURLsWithUser_TransactionError(t *testing.T) {
	mock, repo := setupUserURLsMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	userID := "test-user-id"
	urls := []*model.URLsModel{
		{ShortURL: "batch1", LongURL: "https://example.com/batch1"},
	}
	expectedErr := repository.ErrURLNotFound

	// Ожидаем ошибку при начале транзакции
	mock.ExpectBegin().WillReturnError(expectedErr)

	err := repo.CreateMultipleURLsWithUser(ctx, urls, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}
