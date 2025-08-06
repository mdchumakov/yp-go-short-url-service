package urlExtractor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLinkExtractorService(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис через тестовый конструктор
	service := NewLinkExtractorService(mockRepo)

	// Проверяем, что сервис создан корректно
	assert.NotNil(t, service)
	assert.IsType(t, &linkExtractorService{}, service)
}

func Test_linkExtractorService_ExtractLongURL(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &linkExtractorService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	// Тестовые данные
	shortURL := "abc123"
	longURL := "https://example.com/very/long/url"
	testURL := &model.URLsModel{
		ID:        1,
		ShortURL:  shortURL,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("successful extraction", func(t *testing.T) {
		// Ожидаем успешный вызов GetByShortURL
		mockRepo.EXPECT().
			GetByShortURL(ctx, shortURL).
			Return(testURL, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, shortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, longURL, result)
	})

	t.Run("short URL not found", func(t *testing.T) {
		// Ожидаем вызов GetByShortURL с nil результатом
		mockRepo.EXPECT().
			GetByShortURL(ctx, shortURL).
			Return(nil, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, shortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("database error", func(t *testing.T) {
		expectedErr := errors.New("database connection failed")

		// Ожидаем вызов GetByShortURL с ошибкой
		mockRepo.EXPECT().
			GetByShortURL(ctx, shortURL).
			Return(nil, expectedErr)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, shortURL)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty short URL", func(t *testing.T) {
		emptyShortURL := ""

		// Ожидаем вызов GetByShortURL с пустой строкой
		mockRepo.EXPECT().
			GetByShortURL(ctx, emptyShortURL).
			Return(nil, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, emptyShortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("context without logger", func(t *testing.T) {
		// Создаем контекст без логгера
		ctxWithoutLogger := context.Background()

		// Ожидаем вызов GetByShortURL
		mockRepo.EXPECT().
			GetByShortURL(ctxWithoutLogger, shortURL).
			Return(testURL, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctxWithoutLogger, shortURL)

		// Проверяем результат - должен работать даже без логгера
		assert.NoError(t, err)
		assert.Equal(t, longURL, result)
	})
}

func Test_linkExtractorService_ExtractLongURL_EdgeCases(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &linkExtractorService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	t.Run("very long short URL", func(t *testing.T) {
		veryLongShortURL := strings.Repeat("a", 1000)

		// Ожидаем вызов GetByShortURL с очень длинной строкой
		mockRepo.EXPECT().
			GetByShortURL(ctx, veryLongShortURL).
			Return(nil, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, veryLongShortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("special characters in short URL", func(t *testing.T) {
		specialShortURL := "abc-123_456.789"

		// Ожидаем вызов GetByShortURL с специальными символами
		mockRepo.EXPECT().
			GetByShortURL(ctx, specialShortURL).
			Return(nil, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, specialShortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("URL with empty long URL field", func(t *testing.T) {
		shortURL := "test123"
		testURLWithEmptyLong := &model.URLsModel{
			ID:        1,
			ShortURL:  shortURL,
			LongURL:   "", // Пустая длинная ссылка
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Ожидаем вызов GetByShortURL
		mockRepo.EXPECT().
			GetByShortURL(ctx, shortURL).
			Return(testURLWithEmptyLong, nil)

		// Вызываем метод
		result, err := service.ExtractLongURL(ctx, shortURL)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func Benchmark_linkExtractorService_ExtractLongURL(b *testing.B) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &linkExtractorService{
		repository: mockRepo,
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	// Тестовые данные
	shortURL := "abc123"
	longURL := "https://example.com/very/long/url"
	testURL := &model.URLsModel{
		ID:        1,
		ShortURL:  shortURL,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Настраиваем ожидания мока
	mockRepo.EXPECT().
		GetByShortURL(ctx, shortURL).
		Return(testURL, nil).
		AnyTimes()

	// Запускаем benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ExtractLongURL(ctx, shortURL)
	}
}
