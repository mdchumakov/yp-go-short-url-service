package extractor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"yp-go-short-url-service/internal/middleware"
	serviceMock "yp-go-short-url-service/internal/service/mock"
)

func TestExtractingLongURLHandler_Handle(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name           string
		shortURL       string
		setupMock      func(service *serviceMock.MockURLExtractorService)
		expectedStatus int
		expectedBody   string
		expectedHeader string
		expectedValue  string
	}{
		{
			name:     "успешное перенаправление на длинный URL",
			shortURL: "abc123",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				mockService.EXPECT().
					ExtractLongURL(gomock.Any(), "abc123").
					Return("https://example.com/very/long/url", nil)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "Location",
			expectedValue:  "https://example.com/very/long/url",
		},
		{
			name:     "пустой параметр shortURL",
			shortURL: "",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				// Сервис не должен вызываться при пустом параметре
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Параметр shortURL не может быть пустым",
		},
		{
			name:     "ошибка сервиса при извлечении URL",
			shortURL: "error123",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				mockService.EXPECT().
					ExtractLongURL(gomock.Any(), "error123").
					Return("", errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Ошибка при извлечении длинной ссылки: database connection failed",
		},
		{
			name:     "ссылка не найдена (пустой результат)",
			shortURL: "notfound",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				mockService.EXPECT().
					ExtractLongURL(gomock.Any(), "notfound").
					Return("", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Ссылка не найдена",
		},
		{
			name:     "ссылка не найдена (пустая строка)",
			shortURL: "empty",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				mockService.EXPECT().
					ExtractLongURL(gomock.Any(), "empty").
					Return("", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Ссылка не найдена",
		},
		{
			name:     "специальные символы в shortURL",
			shortURL: "test-123_456",
			setupMock: func(mockService *serviceMock.MockURLExtractorService) {
				mockService.EXPECT().
					ExtractLongURL(gomock.Any(), "test-123_456").
					Return("https://example.com/special-chars", nil)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "Location",
			expectedValue:  "https://example.com/special-chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			gin.SetMode(gin.TestMode)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := serviceMock.NewMockURLExtractorService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewExtractingFullLinkHandler(mockService)

			// Создаем тестовый контекст с логгером
			ctx := context.Background()
			ctx = middleware.WithLogger(ctx, logger)
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

			// Создаем Gin контекст
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			c.Request = c.Request.WithContext(ctx)

			// Устанавливаем Request ID в Gin контексте
			c.Set(middleware.GinRequestIDKey, "test-request-id")

			// Устанавливаем параметр shortURL
			if tt.shortURL != "" {
				c.Params = gin.Params{gin.Param{Key: "shortURL", Value: tt.shortURL}}
			}

			// Act
			handler.Handle(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedValue, w.Header().Get(tt.expectedHeader))
			}
		})
	}
}

func TestExtractingLongURLHandler_Handle_WithRequestID(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("с пользовательским Request ID", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), "custom123").
			Return("https://example.com/custom", nil)

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст с логгером и пользовательским Request ID
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "custom-request-123")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/custom123", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "custom-request-123")
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: "custom123"}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/custom", w.Header().Get("Location"))
	})

	t.Run("без Request ID", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), "no-request-id").
			Return("https://example.com/no-request-id", nil)

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст только с логгером (без Request ID)
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/no-request-id", nil)
		c.Request = c.Request.WithContext(ctx)
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: "no-request-id"}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/no-request-id", w.Header().Get("Location"))
	})
}

func TestExtractingLongURLHandler_Handle_ServiceIntegration(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("проверка вызова сервиса с правильным контекстом", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)

		// Ожидаем, что ExtractLongURL будет вызван с контекстом, содержащим логгер и request ID
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), "test123").
			DoAndReturn(func(ctx context.Context, shortURL string) (string, error) {
				// Проверяем, что в контексте есть логгер
				log := middleware.GetLogger(ctx)
				assert.NotNil(t, log)

				// Проверяем, что в контексте есть request ID
				requestID := middleware.ExtractRequestID(ctx)
				assert.Equal(t, "test-request-id", requestID)

				// Проверяем, что shortURL передается правильно
				assert.Equal(t, "test123", shortURL)

				return "https://example.com/test", nil
			})

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test123", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: "test123"}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/test", w.Header().Get("Location"))
	})
}

func TestNewExtractingFullLinkHandler(t *testing.T) {
	t.Run("создание обработчика с сервисом", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)

		// Act
		handler := NewExtractingFullLinkHandler(mockService)

		// Assert
		assert.NotNil(t, handler)

		// Проверяем, что это правильный тип
		extractingHandler, ok := handler.(*extractingLongURLHandler)
		assert.True(t, ok)
		assert.Equal(t, mockService, extractingHandler.service)
	})
}

func TestExtractingLongURLHandler_Handle_EdgeCases(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("очень длинный shortURL", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		longShortURL := "very-long-short-url-that-exceeds-normal-length-limits-and-should-be-handled-properly"
		mockService := serviceMock.NewMockURLExtractorService(ctrl)
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), longShortURL).
			Return("https://example.com/very-long", nil)

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/"+longShortURL, nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: longShortURL}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/very-long", w.Header().Get("Location"))
	})

	t.Run("специальные символы в URL", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), "special-123_456").
			Return("https://example.com/special-chars", nil)

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/special-123_456", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: "special-123_456"}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://example.com/special-chars", w.Header().Get("Location"))
	})

	t.Run("различные типы ошибок сервиса", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockURLExtractorService(ctrl)
		mockService.EXPECT().
			ExtractLongURL(gomock.Any(), "timeout").
			Return("", errors.New("connection timeout"))

		handler := NewExtractingFullLinkHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/timeout", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")
		c.Params = gin.Params{gin.Param{Key: "shortURL", Value: "timeout"}}

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Ошибка при извлечении длинной ссылки: connection timeout", w.Body.String())
	})
}
