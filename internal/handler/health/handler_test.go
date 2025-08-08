package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"yp-go-short-url-service/internal/middleware"
	serviceMock "yp-go-short-url-service/internal/service/mock"
)

func TestPingHandler_Handle(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name           string
		setupMock      func(*serviceMock.MockHealthCheckService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешная проверка здоровья сервиса",
			setupMock: func(mockService *serviceMock.MockHealthCheckService) {
				mockService.EXPECT().
					Ping(gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
		{
			name: "ошибка проверки здоровья сервиса",
			setupMock: func(mockService *serviceMock.MockHealthCheckService) {
				mockService.EXPECT().
					Ping(gomock.Any()).
					Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "health check failed"}`,
		},
		{
			name: "ошибка с пустым сообщением",
			setupMock: func(mockService *serviceMock.MockHealthCheckService) {
				mockService.EXPECT().
					Ping(gomock.Any()).
					Return(errors.New(""))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": "health check failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			gin.SetMode(gin.TestMode)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := serviceMock.NewMockHealthCheckService(ctrl)
			tt.setupMock(mockService)

			handler := NewPingHandler(mockService)

			// Создаем тестовый контекст с логгером
			ctx := context.Background()
			ctx = middleware.WithLogger(ctx, logger)
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

			// Создаем Gin контекст
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
			c.Request = c.Request.WithContext(ctx)

			// Устанавливаем Request ID в Gin контексте
			c.Set(middleware.GinRequestIDKey, "test-request-id")

			// Act
			handler.Handle(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestPingHandler_Handle_WithRequestID(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("с пользовательским Request ID", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockHealthCheckService(ctrl)
		mockService.EXPECT().
			Ping(gomock.Any()).
			Return(nil)

		handler := NewPingHandler(mockService)

		// Создаем тестовый контекст с логгером и пользовательским Request ID
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "custom-request-123")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "custom-request-123")

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
	})

	t.Run("без Request ID", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockHealthCheckService(ctrl)
		mockService.EXPECT().
			Ping(gomock.Any()).
			Return(nil)

		handler := NewPingHandler(mockService)

		// Создаем тестовый контекст только с логгером (без Request ID)
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		c.Request = c.Request.WithContext(ctx)

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
	})
}

func TestPingHandler_Handle_ServiceIntegration(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("проверка вызова сервиса с правильным контекстом", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockHealthCheckService(ctrl)

		// Ожидаем, что Ping будет вызван с контекстом, содержащим логгер и request ID
		mockService.EXPECT().
			Ping(gomock.Any()).
			DoAndReturn(func(ctx context.Context) error {
				// Проверяем, что в контексте есть логгер
				log := middleware.GetLogger(ctx)
				assert.NotNil(t, log)

				// Проверяем, что в контексте есть request ID
				requestID := middleware.ExtractRequestID(ctx)
				assert.Equal(t, "test-request-id", requestID)

				return nil
			})

		handler := NewPingHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
	})
}

func TestNewPingHandler(t *testing.T) {
	t.Run("создание обработчика с сервисом", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockHealthCheckService(ctrl)

		// Act
		handler := NewPingHandler(mockService)

		// Assert
		assert.NotNil(t, handler)

		// Проверяем, что это правильный тип
		pingHandler, ok := handler.(*pingHandler)
		assert.True(t, ok)
		assert.Equal(t, mockService, pingHandler.service)
	})
}

func TestPingHandler_Handle_EdgeCases(t *testing.T) {
	// Настройка тестового логгера
	logger := zaptest.NewLogger(t).Sugar()

	t.Run("обработка различных типов ошибок", func(t *testing.T) {
		// Arrange
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := serviceMock.NewMockHealthCheckService(ctrl)
		mockService.EXPECT().
			Ping(gomock.Any()).
			Return(errors.New("connection timeout"))

		handler := NewPingHandler(mockService)

		// Создаем тестовый контекст
		ctx := context.Background()
		ctx = middleware.WithLogger(ctx, logger)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")

		// Создаем Gin контекст
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)
		c.Request = c.Request.WithContext(ctx)

		// Устанавливаем Request ID в Gin контексте
		c.Set(middleware.GinRequestIDKey, "test-request-id")

		// Act
		handler.Handle(c)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, `{"error": "health check failed"}`, w.Body.String())
	})
}
