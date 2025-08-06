package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
	"yp-go-short-url-service/internal/repository"

	"yp-go-short-url-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockPool(t *testing.T) (pgxmock.PgxPoolIface, *urlsRepository) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	repo := &urlsRepository{pool: mock}
	return mock, repo
}

func TestNewURLsRepository(t *testing.T) {
	mock, _ := setupMockPool(t)
	defer mock.Close()

	// Создаем репозиторий напрямую с моком
	repo := &urlsRepository{pool: mock}
	assert.NotNil(t, repo)
	assert.IsType(t, &urlsRepository{}, repo)
}

func TestURLsRepository_Ping_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	mock.ExpectExec("SELECT 1").
		WillReturnResult(pgxmock.NewResult("SELECT", 1))

	err := repo.Ping(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_Ping_Error(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedErr := errors.New("database connection error")

	mock.ExpectExec("SELECT 1").
		WillReturnError(expectedErr)

	err := repo.Ping(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByLongURL_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedURL := &model.URLsModel{
		ID:        1,
		ShortURL:  "abc123",
		LongURL:   "https://example.com/very/long/url",
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"}).
		AddRow(expectedURL.ID, expectedURL.ShortURL, expectedURL.LongURL, expectedURL.CreatedAt, expectedURL.UpdatedAt)

	mock.ExpectQuery("SELECT \\* FROM urls WHERE long_url = \\$1").
		WithArgs(expectedURL.LongURL).
		WillReturnRows(rows)

	result, err := repo.GetByLongURL(ctx, expectedURL.LongURL)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURL.ID, result.ID)
	assert.Equal(t, expectedURL.ShortURL, result.ShortURL)
	assert.Equal(t, expectedURL.LongURL, result.LongURL)
	assert.Equal(t, expectedURL.CreatedAt, result.CreatedAt)
	assert.Equal(t, expectedURL.UpdatedAt, result.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByLongURL_NotFound(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	longURL := "https://example.com/not/found"

	mock.ExpectQuery("SELECT \\* FROM urls WHERE long_url = \\$1").
		WithArgs(longURL).
		WillReturnError(pgx.ErrNoRows)

	result, err := repo.GetByLongURL(ctx, longURL)
	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByLongURL_DatabaseError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	longURL := "https://example.com/error"
	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT \\* FROM urls WHERE long_url = \\$1").
		WithArgs(longURL).
		WillReturnError(expectedErr)

	result, err := repo.GetByLongURL(ctx, longURL)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByShortURL_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedURL := &model.URLsModel{
		ID:        1,
		ShortURL:  "abc123",
		LongURL:   "https://example.com/very/long/url",
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"}).
		AddRow(expectedURL.ID, expectedURL.ShortURL, expectedURL.LongURL, expectedURL.CreatedAt, expectedURL.UpdatedAt)

	mock.ExpectQuery("SELECT \\* FROM urls WHERE short_url = \\$1").
		WithArgs(expectedURL.ShortURL).
		WillReturnRows(rows)

	result, err := repo.GetByShortURL(ctx, expectedURL.ShortURL)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURL.ID, result.ID)
	assert.Equal(t, expectedURL.ShortURL, result.ShortURL)
	assert.Equal(t, expectedURL.LongURL, result.LongURL)
	assert.Equal(t, expectedURL.CreatedAt, result.CreatedAt)
	assert.Equal(t, expectedURL.UpdatedAt, result.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByShortURL_NotFound(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	shortURL := "notfound"

	mock.ExpectQuery("SELECT \\* FROM urls WHERE short_url = \\$1").
		WithArgs(shortURL).
		WillReturnError(pgx.ErrNoRows)

	result, err := repo.GetByShortURL(ctx, shortURL)
	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetByShortURL_DatabaseError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	shortURL := "error"
	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT \\* FROM urls WHERE short_url = \\$1").
		WithArgs(shortURL).
		WillReturnError(expectedErr)

	result, err := repo.GetByShortURL(ctx, shortURL)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_Create_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com/very/long/url",
	}

	mock.ExpectExec("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Create(ctx, url)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_Create_DuplicateKeyError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com/very/long/url",
	}

	// Симулируем ошибку дублирования ключа
	pgErr := &pgconn.PgError{
		Code: "23505", // unique_violation
	}

	mock.ExpectExec("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnError(pgErr)

	err := repo.Create(ctx, url)
	assert.Error(t, err)
	assert.Equal(t, pgErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_Create_DatabaseError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com/very/long/url",
	}
	expectedErr := errors.New("database connection error")

	mock.ExpectExec("INSERT INTO urls \\(short_url, long_url\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(url.ShortURL, url.LongURL).
		WillReturnError(expectedErr)

	err := repo.Create(ctx, url)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_Create_WithNilURL(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	// Ожидаем ошибку при передаче nil URL
	err := repo.Create(ctx, nil)
	assert.Error(t, err)
	assert.Equal(t, "url cannot be nil", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Тест для проверки интерфейса
func TestURLsRepository_ImplementsInterface(t *testing.T) {
	mock, _ := setupMockPool(t)
	defer mock.Close()

	// Создаем репозиторий напрямую с моком для проверки интерфейса
	repo := &urlsRepository{pool: mock}
	var _ repository.URLRepository = repo
	var _ repository.URLRepositoryReader = repo
	var _ repository.URLRepositoryWriter = repo
}

// Тест для конструктора
func TestNewURLsRepository_WithRealPool(t *testing.T) {
	// Создаем мок пула
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Создаем репозиторий напрямую, так как NewURLsRepository ожидает *pgxpool.Pool
	repo := &urlsRepository{pool: mock}
	assert.NotNil(t, repo)
	assert.IsType(t, &urlsRepository{}, repo)
}

// Тесты для функций обработки ошибок
func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "pgx.ErrNoRows should return true",
			err:      pgx.ErrNoRows,
			expected: true,
		},
		{
			name:     "ErrURLNotFound should return true",
			err:      repository.ErrURLNotFound,
			expected: true,
		},
		{
			name:     "other error should return false",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error should return false",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repository.IsNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExistsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrURLExists should return true",
			err:      repository.ErrURLExists,
			expected: true,
		},
		{
			name:     "other error should return false",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error should return false",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repository.IsExistsError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
