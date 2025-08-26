package shortener

import (
	"context"
	"errors"
	"testing"
	"time"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/mock"
	services "yp-go-short-url-service/internal/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLinkShortenerService(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	// Создаем сервис через тестовый конструктор
	service := NewURLShortenerService(mockURLRepo, mockUserURLsRepo)

	// Проверяем, что сервис создан корректно
	assert.NotNil(t, service)
	assert.IsType(t, &urlShortenerService{}, service)
}

func Test_linkShortenerService_ShortURL(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
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

		// Проверяем результат - теперь должна возвращаться ошибка
		assert.Error(t, err)
		assert.Equal(t, services.ErrURLAlreadyExists, err)
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
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
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
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
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
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
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

func Test_urlShortenerService_ShortURLsByBatch(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	t.Run("успешное пакетное сокращение - все URL новые", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/url1"},
			{"correlation_id": "2", "original_url": "https://example.com/url2"},
			{"correlation_id": "3", "original_url": "https://example.com/url3"},
		}

		// Ожидаем вызовы GetByLongURL для каждого URL (все не найдены)
		for _, urlData := range longURLs {
			mockRepo.EXPECT().
				GetByLongURL(ctx, urlData["original_url"]).
				Return(nil, repository.ErrURLNotFound)
		}

		// Ожидаем вызов CreateBatch для сохранения новых URL
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, urls []*model.URLsModel) error {
				// Проверяем, что передано правильное количество URL
				assert.Len(t, urls, 3)

				// Проверяем каждый URL
				for i, url := range urls {
					assert.Equal(t, longURLs[i]["original_url"], url.LongURL)
					assert.NotEmpty(t, url.ShortURL)
					assert.Len(t, url.ShortURL, 8) // Проверяем длину short URL
					assert.NotZero(t, url.CreatedAt)
					assert.NotZero(t, url.UpdatedAt)
				}
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)

		// Проверяем, что в результате есть short_url для каждого URL
		for i, urlData := range result {
			assert.Equal(t, longURLs[i]["correlation_id"], urlData["correlation_id"])
			assert.Equal(t, longURLs[i]["original_url"], urlData["original_url"])
			assert.NotEmpty(t, urlData["short_url"])
			assert.Len(t, urlData["short_url"], 8)
		}
	})

	t.Run("пакетное сокращение - некоторые URL уже существуют", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/existing1"},
			{"correlation_id": "2", "original_url": "https://example.com/new2"},
			{"correlation_id": "3", "original_url": "https://example.com/existing3"},
		}

		// Существующие URL
		existingURL1 := &model.URLsModel{
			ID:        1,
			ShortURL:  "existing1",
			LongURL:   "https://example.com/existing1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		existingURL3 := &model.URLsModel{
			ID:        3,
			ShortURL:  "existing3",
			LongURL:   "https://example.com/existing3",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Ожидаем вызовы GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/existing1").
			Return(existingURL1, nil)
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/new2").
			Return(nil, repository.ErrURLNotFound)
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/existing3").
			Return(existingURL3, nil)

		// Ожидаем вызов CreateBatch только для нового URL
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, urls []*model.URLsModel) error {
				// Проверяем, что передано только один новый URL
				assert.Len(t, urls, 1)
				assert.Equal(t, "https://example.com/new2", urls[0].LongURL)
				assert.NotEmpty(t, urls[0].ShortURL)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)

		// Проверяем, что существующие URL имеют правильные short_url
		assert.Equal(t, "existing1", result[0]["short_url"])
		assert.NotEmpty(t, result[1]["short_url"]) // Новый URL
		assert.Equal(t, "existing3", result[2]["short_url"])
	})

	t.Run("ошибка при поиске существующего URL", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/url1"},
		}

		expectedErr := errors.New("database connection failed")

		// Ожидаем вызов GetByLongURL с ошибкой
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/url1").
			Return(nil, expectedErr)

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("ошибка при сохранении в базу данных", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/url1"},
			{"correlation_id": "2", "original_url": "https://example.com/url2"},
		}

		expectedErr := errors.New("database write failed")

		// Ожидаем вызовы GetByLongURL (все не найдены)
		for _, urlData := range longURLs {
			mockRepo.EXPECT().
				GetByLongURL(ctx, urlData["original_url"]).
				Return(nil, repository.ErrURLNotFound)
		}

		// Ожидаем вызов CreateBatch с ошибкой
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			Return(expectedErr)

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("пустой массив URL", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{}

		// Ожидаем вызов CreateBatch с пустым массивом
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, urls []*model.URLsModel) error {
				assert.Len(t, urls, 0)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("один URL в пакете", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/single"},
		}

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/single").
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов CreateBatch
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, urls []*model.URLsModel) error {
				assert.Len(t, urls, 1)
				assert.Equal(t, "https://example.com/single", urls[0].LongURL)
				assert.NotEmpty(t, urls[0].ShortURL)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "1", result[0]["correlation_id"])
		assert.Equal(t, "https://example.com/single", result[0]["original_url"])
		assert.NotEmpty(t, result[0]["short_url"])
	})

	t.Run("контекст без логгера", func(t *testing.T) {
		// Создаем контекст без логгера
		ctxWithoutLogger := context.Background()

		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/test"},
		}

		// Ожидаем вызов GetByLongURL
		mockRepo.EXPECT().
			GetByLongURL(ctxWithoutLogger, "https://example.com/test").
			Return(nil, repository.ErrURLNotFound)

		// Ожидаем вызов CreateBatch
		mockRepo.EXPECT().
			CreateBatch(ctxWithoutLogger, gomock.Any()).
			Return(nil)

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctxWithoutLogger, longURLs)

		// Проверяем результат - должен работать даже без логгера
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
	})
}

func Test_urlShortenerService_ShortURL_ConflictScenarios(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepository(ctrl)

	// Создаем сервис
	service := &urlShortenerService{
		urlRepository:      mockRepo,
		userURLsRepository: mock.NewMockUserURLsRepository(ctrl),
	}

	// Создаем контекст с логгером для тестов
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())

	t.Run("конфликт при попытке сократить уже существующий URL", func(t *testing.T) {
		longURL := "https://example.com/existing/url"
		existingShortURL := "existing123"
		existingURL := &model.URLsModel{
			ID:        1,
			ShortURL:  existingShortURL,
			LongURL:   longURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Ожидаем вызов GetByLongURL с существующим URL
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL).
			Return(existingURL, nil)

		// Вызываем метод
		result, err := service.ShortURL(ctx, longURL)

		// Проверяем результат - должна возвращаться ошибка конфликта
		assert.Error(t, err)
		assert.Equal(t, services.ErrURLAlreadyExists, err)
		assert.Equal(t, existingShortURL, result)
	})

	t.Run("конфликт с разными длинными URL, но одинаковыми короткими", func(t *testing.T) {
		longURL1 := "https://example.com/url1"
		longURL2 := "https://example.com/url2"

		// Первый вызов - URL не найден, создаем новый
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL1).
			Return(nil, repository.ErrURLNotFound)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, url *model.URLsModel) error {
				assert.Equal(t, longURL1, url.LongURL)
				assert.NotEmpty(t, url.ShortURL)
				return nil
			})

		result1, err1 := service.ShortURL(ctx, longURL1)
		assert.NoError(t, err1)
		assert.NotEmpty(t, result1)

		// Второй вызов - URL не найден, но при создании возникает конфликт
		mockRepo.EXPECT().
			GetByLongURL(ctx, longURL2).
			Return(nil, repository.ErrURLNotFound)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(repository.ErrURLExists)

		result2, err2 := service.ShortURL(ctx, longURL2)
		assert.Error(t, err2)
		assert.Equal(t, repository.ErrURLExists, err2)
		assert.Equal(t, "", result2)
	})

	t.Run("проверка IsAlreadyExistsError функции", func(t *testing.T) {
		// Тестируем функцию проверки ошибки
		assert.True(t, services.IsAlreadyExistsError(services.ErrURLAlreadyExists))
		assert.False(t, services.IsAlreadyExistsError(errors.New("other error")))
		assert.False(t, services.IsAlreadyExistsError(nil))
	})

	t.Run("конфликт в пакетном режиме - все URL уже существуют", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/existing1"},
			{"correlation_id": "2", "original_url": "https://example.com/existing2"},
		}

		// Существующие URL
		existingURL1 := &model.URLsModel{
			ID:        1,
			ShortURL:  "existing1",
			LongURL:   "https://example.com/existing1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		existingURL2 := &model.URLsModel{
			ID:        2,
			ShortURL:  "existing2",
			LongURL:   "https://example.com/existing2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Ожидаем вызовы GetByLongURL - все URL найдены
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/existing1").
			Return(existingURL1, nil)
		mockRepo.EXPECT().
			GetByLongURL(ctx, "https://example.com/existing2").
			Return(existingURL2, nil)

		// Ожидаем вызов CreateBatch с пустым массивом (нет новых URL для создания)
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, urls []*model.URLsModel) error {
				assert.Len(t, urls, 0)
				return nil
			})

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат - должен вернуть существующие short URL без ошибки
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "existing1", result[0]["short_url"])
		assert.Equal(t, "existing2", result[1]["short_url"])
	})

	t.Run("конфликт в пакетном режиме - ошибка при создании", func(t *testing.T) {
		// Тестовые данные
		longURLs := []map[string]string{
			{"correlation_id": "1", "original_url": "https://example.com/new1"},
			{"correlation_id": "2", "original_url": "https://example.com/new2"},
		}

		// Ожидаем вызовы GetByLongURL - все URL не найдены
		for _, urlData := range longURLs {
			mockRepo.EXPECT().
				GetByLongURL(ctx, urlData["original_url"]).
				Return(nil, repository.ErrURLNotFound)
		}

		// Ожидаем вызов CreateBatch с ошибкой конфликта
		mockRepo.EXPECT().
			CreateBatch(ctx, gomock.Any()).
			Return(repository.ErrURLExists)

		// Вызываем метод
		result, err := service.ShortURLsByBatch(ctx, longURLs)

		// Проверяем результат - должна возвращаться ошибка
		assert.Error(t, err)
		assert.Equal(t, repository.ErrURLExists, err)
		assert.Nil(t, result)
	})
}
