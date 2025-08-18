package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserURLsRepository_GetByUserID(t *testing.T) {
	// Подготовка тестовой базы данных
	db, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES (?, ?, ?)`,
		userID, "testuser", false)
	require.NoError(t, err)

	// Создаем тестовые URL
	url1 := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com/1",
	}
	url2 := &model.URLsModel{
		ShortURL: "def456",
		LongURL:  "https://example.com/2",
	}

	// Связываем URL с пользователем
	err = repo.CreateURLWithUser(ctx, url1, userID)
	require.NoError(t, err)
	err = repo.CreateURLWithUser(ctx, url2, userID)
	require.NoError(t, err)

	// Тестируем получение URL пользователя
	urls, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, urls, 2)

	// Проверяем, что URL содержат правильные данные
	urlMap := make(map[string]*model.URLsModel)
	for _, url := range urls {
		urlMap[url.ShortURL] = url
	}

	assert.Contains(t, urlMap, "abc123")
	assert.Contains(t, urlMap, "def456")
	assert.Equal(t, "https://example.com/1", urlMap["abc123"].LongURL)
	assert.Equal(t, "https://example.com/2", urlMap["def456"].LongURL)
}

func TestUserURLsRepository_AddURL(t *testing.T) {
	// Подготовка тестовой базы данных
	db, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES (?, ?, ?)`,
		userID, "testuser", false)
	require.NoError(t, err)

	// Создаем тестовый URL
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	// Тестируем добавление URL для пользователя
	err = repo.CreateURLWithUser(ctx, url, userID)
	assert.NoError(t, err)

	// Проверяем, что связь создана
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = ? AND url_id = ?`,
		userID, url.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Тестируем повторное добавление (должно вернуть ошибку "уже существует")
	err = repo.CreateURLWithUser(ctx, url, userID)
	assert.Error(t, err)
	assert.True(t, repository.IsExistsError(err))
}

func TestUserURLsRepository_AddURL_Validation(t *testing.T) {
	// Подготовка тестовой базы данных
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(db)
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

func TestUserURLsRepository_CreateURLWithUserBasic(t *testing.T) {
	// Подготовка тестовой базы данных
	db, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES (?, ?, ?)`,
		userID, "testuser", false)
	require.NoError(t, err)

	// Тестируем создание URL с пользователем
	url := &model.URLsModel{
		ShortURL: "basic123",
		LongURL:  "https://example.com/basic",
	}

	err = repo.CreateURLWithUser(ctx, url, userID)
	require.NoError(t, err)

	// Проверяем, что URL создан
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM urls WHERE short_url = ?`, url.ShortURL).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Проверяем связь с пользователем
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = ? AND url_id = ?`,
		userID, url.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Тестируем валидацию
	err = repo.CreateURLWithUser(ctx, nil, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url cannot be nil")

	err = repo.CreateURLWithUser(ctx, &model.URLsModel{ShortURL: "test", LongURL: "test"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")
}

func TestUserURLsRepository_CreateMultipleURLsWithUser(t *testing.T) {
	// Подготовка тестовой базы данных
	db, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES (?, ?, ?)`,
		userID, "testuser", false)
	require.NoError(t, err)

	// Тестируем создание множественных URL
	urls := []*model.URLsModel{
		{ShortURL: "batch1", LongURL: "https://example.com/batch1"},
		{ShortURL: "batch2", LongURL: "https://example.com/batch2"},
		{ShortURL: "batch3", LongURL: "https://example.com/batch3"},
	}

	err = repo.CreateMultipleURLsWithUser(ctx, urls, userID)
	require.NoError(t, err)

	// Проверяем, что все URL созданы и связаны с пользователем
	for _, url := range urls {
		var count int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = ? AND url_id = ?`,
			userID, url.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	}

	// Тестируем пустой список
	err = repo.CreateMultipleURLsWithUser(ctx, []*model.URLsModel{}, userID)
	assert.NoError(t, err)

	// Тестируем валидацию
	err = repo.CreateMultipleURLsWithUser(ctx, urls, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")
}

// setupUserURLsTestDB создает тестовую базу данных и возвращает соединение
func setupUserURLsTestDB(t *testing.T) (*sql.DB, func()) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Создаем таблицы
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			password TEXT,
			is_anonymous BOOLEAN DEFAULT FALSE,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_url TEXT NOT NULL UNIQUE,
			long_url TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			url_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
			UNIQUE(user_id, url_id)
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
