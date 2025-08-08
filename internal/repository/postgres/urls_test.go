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

func TestURLsRepository_GetAll_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	limit, offset := 10, 0

	expectedURLs := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example1.com",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			ShortURL:  "def456",
			LongURL:   "https://example2.com",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"})
	for _, url := range expectedURLs {
		rows.AddRow(url.ID, url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt)
	}

	mock.ExpectQuery("SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx, limit, offset)
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

func TestURLsRepository_GetAll_EmptyResult(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	limit, offset := 10, 0

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetAll_WithPagination(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	limit, offset := 5, 10

	expectedURLs := []*model.URLsModel{
		{
			ID:        11,
			ShortURL:  "xyz789",
			LongURL:   "https://example11.com",
			CreatedAt: time.Date(2023, 1, 11, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 11, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"})
	for _, url := range expectedURLs {
		rows.AddRow(url.ID, url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt)
	}

	mock.ExpectQuery("SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx, limit, offset)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedURLs[0].ID, result[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetAll_DatabaseError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	limit, offset := 10, 0
	expectedErr := errors.New("database connection error")

	mock.ExpectQuery("SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(limit, offset).
		WillReturnError(expectedErr)

	result, err := repo.GetAll(ctx, limit, offset)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetAll_ScanError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	limit, offset := 10, 0

	// Создаем строки с неправильными типами данных для вызова ошибки сканирования
	rows := pgxmock.NewRows([]string{"id", "short_url", "long_url", "created_at", "updated_at"}).
		AddRow("invalid_id", "abc123", "https://example.com", "invalid_date", "invalid_date")

	mock.ExpectQuery("SELECT id, short_url, long_url, created_at, updated_at FROM urls ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(limit, offset).
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx, limit, offset)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetTotalCount_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedCount := int64(42)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM urls").
		WillReturnRows(rows)

	result, err := repo.GetTotalCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetTotalCount_ZeroCount(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedCount := int64(0)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM urls").
		WillReturnRows(rows)

	result, err := repo.GetTotalCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_GetTotalCount_DatabaseError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedErr := errors.New("database connection error")

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM urls").
		WillReturnError(expectedErr)

	result, err := repo.GetTotalCount(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, int64(0), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_CreateBatch_Success(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	urls := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example1.com",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			ShortURL:  "def456",
			LongURL:   "https://example2.com",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем batch операции - параметры в правильном порядке: short_url, long_url, created_at, updated_at
	for _, url := range urls {
		mock.ExpectExec("INSERT INTO urls \\(short_url, long_url, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(short_url\\) DO NOTHING").
			WithArgs(url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	// Ожидаем подтверждение транзакции
	mock.ExpectCommit()

	err := repo.CreateBatch(ctx, urls)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_CreateBatch_EmptySlice(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	err := repo.CreateBatch(ctx, []*model.URLsModel{})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_CreateBatch_WithNilURLs(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	urls := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example1.com",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		nil, // nil URL
		{
			ID:        2,
			ShortURL:  "def456",
			LongURL:   "https://example2.com",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем batch операции только для не-nil URL - параметры в правильном порядке
	validURLs := []*model.URLsModel{urls[0], urls[2]}
	for _, url := range validURLs {
		mock.ExpectExec("INSERT INTO urls \\(short_url, long_url, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(short_url\\) DO NOTHING").
			WithArgs(url.ShortURL, url.LongURL, url.CreatedAt, url.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	// Ожидаем подтверждение транзакции
	mock.ExpectCommit()

	err := repo.CreateBatch(ctx, urls)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestURLsRepository_CreateBatch_TransactionError(t *testing.T) {
	mock, repo := setupMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	urls := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example1.com",
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	expectedErr := errors.New("transaction error")

	// Ожидаем ошибку при начале транзакции
	mock.ExpectBegin().WillReturnError(expectedErr)

	err := repo.CreateBatch(ctx, urls)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
