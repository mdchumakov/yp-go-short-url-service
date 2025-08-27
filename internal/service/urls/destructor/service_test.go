package destructor

import (
	"context"
	"testing"
	"time"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// contextKey - тип для ключей контекста в тестах
type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

func TestURLDestructorService_DeleteURLsByBatch_Async(t *testing.T) {
	// Создаем простую реализацию репозитория для тестирования
	testRepo := &testUserURLsRepository{
		deletedURLs: make(map[string][]string),
	}

	// Создаем сервис
	service := NewURLDestructorService(nil, testRepo)
	defer service.Stop() // Важно остановить сервис после теста

	// Создаем контекст с логгером и пользователем
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()
	ctx := context.WithValue(context.Background(), loggerKey, sugaredLogger)
	ctx = context.WithValue(ctx, requestIDKey, "test-request-id")

	// Создаем тестового пользователя и добавляем в контекст
	user := &model.UserModel{
		ID: "test-user-id",
	}
	ctx = context.WithValue(ctx, middleware.JWTTokenContextKey, user)

	// Вызываем метод
	err := service.DeleteURLsByBatch(ctx, []string{"url1", "url2"})

	// Проверяем, что ошибки нет (запрос отправлен в канал)
	assert.NoError(t, err)

	// Ждем немного, чтобы горутина успела обработать запрос
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что URLs были удалены
	assert.Contains(t, testRepo.deletedURLs, "test-user-id")
	assert.Equal(t, []string{"url1", "url2"}, testRepo.deletedURLs["test-user-id"])
}

// testUserURLsRepository - простая реализация для тестирования
type testUserURLsRepository struct {
	deletedURLs map[string][]string
}

func (t *testUserURLsRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URLsModel, error) {
	return nil, nil
}

func (t *testUserURLsRepository) CreateURLWithUser(ctx context.Context, url *model.URLsModel, userID string) error {
	return nil
}

func (t *testUserURLsRepository) CreateMultipleURLsWithUser(ctx context.Context, urls []*model.URLsModel, userID string) error {
	return nil
}

func (t *testUserURLsRepository) DeleteURLsWithUser(ctx context.Context, shortURLs []string, userID string) error {
	t.deletedURLs[userID] = shortURLs
	return nil
}

func TestURLDestructorService_Stop(t *testing.T) {
	// Создаем простую реализацию репозитория для тестирования
	testRepo := &testUserURLsRepository{
		deletedURLs: make(map[string][]string),
	}

	// Создаем сервис
	service := NewURLDestructorService(nil, testRepo)

	// Проверяем, что сервис можно остановить без ошибок
	assert.NotPanics(t, func() {
		service.Stop()
	})
}
