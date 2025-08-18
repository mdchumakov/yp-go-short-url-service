package postgres

import (
	"context"
	"testing"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserURLsRepository_GetByUserID(t *testing.T) {
	// Подготовка тестовой базы данных
	pool, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(pool)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := pool.Exec(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES ($1, $2, $3)`,
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

	// Добавляем URL в таблицу urls
	_, err = pool.Exec(ctx, `INSERT INTO urls (short_url, long_url) VALUES ($1, $2)`,
		url1.ShortURL, url1.LongURL)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO urls (short_url, long_url) VALUES ($1, $2)`,
		url2.ShortURL, url2.LongURL)
	require.NoError(t, err)

	// Получаем ID созданных URL
	var url1ID, url2ID uint
	err = pool.QueryRow(ctx, `SELECT id FROM urls WHERE short_url = $1`, url1.ShortURL).Scan(&url1ID)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT id FROM urls WHERE short_url = $1`, url2.ShortURL).Scan(&url2ID)
	require.NoError(t, err)

	url1.ID = url1ID
	url2.ID = url2ID

	// Связываем URL с пользователем
	err = repo.AddURL(ctx, userID, url1)
	require.NoError(t, err)
	err = repo.AddURL(ctx, userID, url2)
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
	pool, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(pool)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := pool.Exec(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES ($1, $2, $3)`,
		userID, "testuser", false)
	require.NoError(t, err)

	// Создаем тестовый URL
	url := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com",
	}

	// Добавляем URL в таблицу urls
	_, err = pool.Exec(ctx, `INSERT INTO urls (short_url, long_url) VALUES ($1, $2)`,
		url.ShortURL, url.LongURL)
	require.NoError(t, err)

	// Получаем ID созданного URL
	var urlID uint
	err = pool.QueryRow(ctx, `SELECT id FROM urls WHERE short_url = $1`, url.ShortURL).Scan(&urlID)
	require.NoError(t, err)
	url.ID = urlID

	// Тестируем добавление URL для пользователя
	err = repo.AddURL(ctx, userID, url)
	assert.NoError(t, err)

	// Проверяем, что связь создана
	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = $1 AND url_id = $2`,
		userID, url.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Тестируем повторное добавление (должно вернуть ошибку)
	err = repo.AddURL(ctx, userID, url)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrURLExists, err)
}

func TestUserURLsRepository_AddURL_Validation(t *testing.T) {
	// Подготовка тестовой базы данных
	pool, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(pool)
	ctx := context.Background()

	// Тестируем валидацию пустого userID
	err := repo.AddURL(ctx, "", &model.URLsModel{ID: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")

	// Тестируем валидацию nil URL
	err = repo.AddURL(ctx, "test-user-id", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url cannot be nil")
}

func TestUserURLsRepository_CreateURLWithUserBasic(t *testing.T) {
	// Подготовка тестовой базы данных
	pool, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(pool)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := pool.Exec(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES ($1, $2, $3)`,
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
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM urls WHERE short_url = $1`, url.ShortURL).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Проверяем связь с пользователем
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = $1 AND url_id = $2`,
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
	pool, cleanup := setupUserURLsTestDB(t)
	defer cleanup()

	repo := NewUserURLsRepository(pool)
	ctx := context.Background()

	// Создаем тестового пользователя
	userID := "test-user-id"
	_, err := pool.Exec(ctx, `INSERT INTO users (id, name, is_anonymous) VALUES ($1, $2, $3)`,
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
		err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_urls WHERE user_id = $1 AND url_id = $2`,
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

// setupUserURLsTestDB создает тестовую базу данных и возвращает пул соединений
func setupUserURLsTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Здесь должна быть логика создания тестовой базы данных
	// Для простоты используем существующую базу
	pool, err := pgxpool.New(context.Background(), "postgres://localhost:5432/test_db")
	require.NoError(t, err)

	// Очистка таблиц перед тестом
	ctx := context.Background()
	_, err = pool.Exec(ctx, `DELETE FROM user_urls`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM urls`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM users`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}
