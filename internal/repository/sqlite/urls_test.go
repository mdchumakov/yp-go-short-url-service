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

func TestURLsRepository_GetAll_Success(t *testing.T) {
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

	// Получаем все URL с пагинацией
	result, err := repo.GetAll(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Проверяем, что все URL получены (порядок может быть разным из-за сортировки по created_at DESC)
	shortURLs := make(map[string]string)
	for _, url := range result {
		shortURLs[url.ShortURL] = url.LongURL
	}

	assert.Equal(t, "https://example1.com", shortURLs["abc123"])
	assert.Equal(t, "https://example2.com", shortURLs["def456"])
	assert.Equal(t, "https://example3.com", shortURLs["ghi789"])
}

func TestURLsRepository_GetAll_EmptyResult(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Получаем все URL из пустой базы
	result, err := repo.GetAll(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestURLsRepository_GetAll_WithPagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем 5 URL
	urls := []*model.URLsModel{
		{ShortURL: "abc123", LongURL: "https://example1.com"},
		{ShortURL: "def456", LongURL: "https://example2.com"},
		{ShortURL: "ghi789", LongURL: "https://example3.com"},
		{ShortURL: "jkl012", LongURL: "https://example4.com"},
		{ShortURL: "mno345", LongURL: "https://example5.com"},
	}

	for _, url := range urls {
		err := repo.Create(context.Background(), url)
		require.NoError(t, err)
	}

	// Получаем первые 2 записи
	result, err := repo.GetAll(context.Background(), 2, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Получаем следующие 2 записи
	result2, err := repo.GetAll(context.Background(), 2, 2)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Len(t, result2, 2)

	// Получаем последнюю запись
	result3, err := repo.GetAll(context.Background(), 2, 4)
	assert.NoError(t, err)
	assert.NotNil(t, result3)
	assert.Len(t, result3, 1)

	// Проверяем, что все записи разные
	allShortURLs := make(map[string]bool)
	for _, url := range result {
		allShortURLs[url.ShortURL] = true
	}
	for _, url := range result2 {
		allShortURLs[url.ShortURL] = true
	}
	for _, url := range result3 {
		allShortURLs[url.ShortURL] = true
	}

	assert.Len(t, allShortURLs, 5) // Все 5 записей должны быть уникальными
}

func TestURLsRepository_GetAll_WithLargeOffset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем 3 URL
	urls := []*model.URLsModel{
		{ShortURL: "abc123", LongURL: "https://example1.com"},
		{ShortURL: "def456", LongURL: "https://example2.com"},
		{ShortURL: "ghi789", LongURL: "https://example3.com"},
	}

	for _, url := range urls {
		err := repo.Create(context.Background(), url)
		require.NoError(t, err)
	}

	// Пытаемся получить записи с большим offset
	result, err := repo.GetAll(context.Background(), 10, 100)
	assert.NoError(t, err)
	assert.Len(t, result, 0) // Должен вернуть пустой результат
}

func TestURLsRepository_GetTotalCount_Success(t *testing.T) {
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

	// Получаем общее количество
	count, err := repo.GetTotalCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestURLsRepository_GetTotalCount_EmptyDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Получаем количество из пустой базы
	count, err := repo.GetTotalCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestURLsRepository_GetAllAndGetTotalCount_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewURLsRepository(db)

	// Создаем 10 URL
	urls := []*model.URLsModel{
		{ShortURL: "abc123", LongURL: "https://example1.com"},
		{ShortURL: "def456", LongURL: "https://example2.com"},
		{ShortURL: "ghi789", LongURL: "https://example3.com"},
		{ShortURL: "jkl012", LongURL: "https://example4.com"},
		{ShortURL: "mno345", LongURL: "https://example5.com"},
		{ShortURL: "pqr678", LongURL: "https://example6.com"},
		{ShortURL: "stu901", LongURL: "https://example7.com"},
		{ShortURL: "vwx234", LongURL: "https://example8.com"},
		{ShortURL: "yza567", LongURL: "https://example9.com"},
		{ShortURL: "bcd890", LongURL: "https://example10.com"},
	}

	for _, url := range urls {
		err := repo.Create(context.Background(), url)
		require.NoError(t, err)
	}

	// Получаем общее количество
	totalCount, err := repo.GetTotalCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(10), totalCount)

	// Получаем все записи постранично
	limit := 3
	totalRetrieved := 0
	offset := 0

	for {
		result, err := repo.GetAll(context.Background(), limit, offset)
		assert.NoError(t, err)

		if len(result) == 0 {
			break
		}

		totalRetrieved += len(result)
		offset += limit
	}

	// Проверяем, что получили все записи
	assert.Equal(t, int(totalCount), totalRetrieved)
}
