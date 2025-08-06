package urlShortener

import (
	"context"
	"errors"
	"testing"
	"time"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLinkShortenerService(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис через тестовый конструктор
	service := NewLinkShortenerService(mockRepo)

	// Проверяем, что сервис создан корректно
	assert.NotNil(t, service)
	assert.IsType(t, &linkShortenerService{}, service)
}

func Test_linkShortenerService_ShortURL(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &linkShortenerService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	// Тестовые данные
	longURL := "https://example.com/very/long/url/that/needs/to/be/shortened"
	existingShortURL := "abc123"
	existingURL := &model.URLsModel{
		ID:        1,
		ShortURL:  existingShortURL,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("successful shortening - new URL", func(t *testing.T) {
		// Ожидаем вызов GetByLongURL с nil результатом (URL не найден)
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create для сохранения нового URL
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				// Проверяем, что URL содержит правильные данные
				assert.Equal(t, longURL, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				assert.Len(t, url.ShortURL, 8) // Проверяем длину short URL
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURL(ctx, longURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 8)
	})

	t.Run("successful shortening - URL already exists", func(t *testing.T) {
		// Ожидаем вызов GetByLongURL с существующим URL
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(existingURL, nil)

		// Вызываем метод
		result, err := service.ShortURL(ctx, longURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, existingShortURL, result)
	})

	t.Run("database error during search", func(t *testing.T) {
		expectedErr := errors.New("database connection failed")

		// Ожидаем вызов GetByLongURL с ошибкой
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(nil, expectedErr)

		// Вызываем метод
		result, err := service.ShortURL(ctx, longURL)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", result)
	})

	t.Run("database error during save", func(t *testing.T) {
		expectedErr := errors.New("database write failed")

		// Ожидаем вызов GetByLongURL с nil результатом
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create с ошибкой
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(expectedErr)

		// Вызываем метод
		result, err := service.ShortURL(ctx, longURL)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty long URL", func(t *testing.T) {
		emptyLongURL := ""

		// Ожидаем вызов GetByLongURL с пустой строкой
		mockRepo.EXPECT().
			GetByLongURL(ctx, emptyLongURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create для сохранения
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				assert.Equal(t, emptyLongURL, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURL(ctx, emptyLongURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("context without logger", func(t *testing.T) {
		// Создаем контекст без логгера
		ctxWithoutLogger := context.Background()

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctxWithoutLogger, longURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create
		mockRepo.EXPECT().
			Create(ctxWithoutLogger, gomock.Any()).
			Return(nil)

		// Вызываем метод
		result, err := service.ShortURL(ctxWithoutLogger, longURL)

		// Проверяем результат - должен работать даже без логгера
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func Test_linkShortenerService_ShortURL_EdgeCases(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &linkShortenerService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	t.Run("very long URL", func(t *testing.T) {
		veryLongURL := "https://example.com/" + string(make([]byte, 10000)) // Очень длинный URL

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctx, veryLongURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				assert.Equal(t, veryLongURL, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				assert.Len(t, url.ShortURL, 8)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURL(ctx, veryLongURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 8)
	})

	t.Run("URL with special characters", func(t *testing.T) {
		specialURL := "https://example.com/path?param=value&another=param#fragment"

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctx, specialURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				assert.Equal(t, specialURL, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURL(ctx, specialURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("URL with unicode characters", func(t *testing.T) {
		unicodeURL := "https://example.com/путь/с/русскими/символами"

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctx, unicodeURL).
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов Create
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				assert.Equal(t, unicodeURL, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURL(ctx, unicodeURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func Test_linkShortenerService_extractShortURLIfExists(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &linkShortenerService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	longURL := "https://example.com/test"
	existingURL := &model.URLsModel{
		ID:        1,
		ShortURL:  "abc123",
		LongURL:   longURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("URL found", func(t *testing.T) {
		// Ожидаем вызов GetByLongURL с существующим URL
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(existingURL, nil)

		// Вызываем метод
		result, err := service.extractShortURLIfExists(ctx, longURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, existingURL.ShortURL, *result)
	})

	t.Run("URL not found", func(t *testing.T) {
		// Ожидаем вызов GetByLongURL с ошибкой "не найдено"
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(nil, repository.ErrURLNotFound)

		// Вызываем метод
		result, err := service.extractShortURLIfExists(ctx, longURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("database error", func(t *testing.T) {
		expectedErr := errors.New("database error")

		// Ожидаем вызов GetByLongURL с ошибкой
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(nil, expectedErr)

		// Вызываем метод
		result, err := service.extractShortURLIfExists(ctx, longURL)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func Test_linkShortenerService_saveShortURLToStorage(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &linkShortenerService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	testURL := &model.URLsModel{
		ShortURL: "abc123",
		LongURL:  "https://example.com/test",
	}

	t.Run("successful save", func(t *testing.T) {
		// Ожидаем успешный вызов Create
		mockRepo.EXPECT().
			Create(ctx, testURL).
			Return(nil)

		// Вызываем метод
		err := service.saveShortURLToStorage(ctx, testURL)

		// Проверяем результат
		assert.NoError(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		expectedErr := errors.New("save failed")

		// Ожидаем вызов Create с ошибкой
		mockRepo.EXPECT().
			Create(ctx, testURL).
			Return(expectedErr)

		// Вызываем метод
		err := service.saveShortURLToStorage(ctx, testURL)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
