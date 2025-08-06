package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Создаем таблицу для тестов
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_url TEXT NOT NULL UNIQUE,
		long_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestNewURLsRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)
	assert.NotNil(t, repo)
}

func TestURLsRepository_Ping(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	err := repo.Ping(context.Background())
	assert.NoError(t, err)
}

func TestURLsRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	err := repo.Create(context.Background(), url)
	assert.NoError(t, err)
}

func TestURLsRepository_Create_NilURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	err := repo.Create(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url cannot be nil")
}

func TestURLsRepository_GetByLongURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем тестовые данные
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	err := repo.Create(context.Background(), url)
	require.NoError(t, err)

	// Получаем URL по длинному URL
	result, err := repo.GetByLongURL(context.Background(), "https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.ShortURL)
	assert.Equal(t, "https://example.com", result.LongURL)
}

func TestURLsRepository_GetByLongURL_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Пытаемся получить несуществующий URL
	result, err := repo.GetByLongURL(context.Background(), "https://nonexistent.com")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrURLNotFound, err)
}

func TestURLsRepository_GetByShortURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем тестовые данные
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	err := repo.Create(context.Background(), url)
	require.NoError(t, err)

	// Получаем URL по короткому URL
	result, err := repo.GetByShortURL(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.ShortURL)
	assert.Equal(t, "https://example.com", result.LongURL)
}

func TestURLsRepository_GetByShortURL_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Пытаемся получить несуществующий URL
	result, err := repo.GetByShortURL(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrURLNotFound, err)
}

func TestURLsRepository_CreateAndRetrieve(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем несколько URL
	urls := []*model.URLsModel{
		{ShortURL: "abc123", LongURL: "https://example1.com"},
		{ShortURL: "def456", LongURL: "https://example2.com"},
		{ShortURL: "ghi789", LongURL: "https://example3.com"},
	}

	for _, url := range urls {
		err := repo.Create(context.Background(), url)
		require.NoError(t, err)
	}

	// Проверяем, что все URL можно получить
	for _, expectedURL := range urls {
		// По короткому URL
		result, err := repo.GetByShortURL(context.Background(), expectedURL.ShortURL)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL.ShortURL, result.ShortURL)
		assert.Equal(t, expectedURL.LongURL, result.LongURL)

		// По длинному URL
		result, err = repo.GetByLongURL(context.Background(), expectedURL.LongURL)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL.ShortURL, result.ShortURL)
		assert.Equal(t, expectedURL.LongURL, result.LongURL)
	}
}

func TestURLsRepository_Timestamps(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	err := repo.Create(context.Background(), url)
	require.NoError(t, err)

	// Получаем созданный URL
	result, err := repo.GetByShortURL(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Проверяем, что временные метки установлены
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
	assert.True(t, result.CreatedAt.Before(time.Now().Add(time.Second)))
	assert.True(t, result.UpdatedAt.Before(time.Now().Add(time.Second)))
}
