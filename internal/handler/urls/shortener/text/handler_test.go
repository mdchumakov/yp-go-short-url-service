package text

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const apiPath = "/"

func getDefaultSettings() *config.Settings {
	return &config.Settings{
		EnvSettings: &config.ENVSettings{
			Server: &config.ServerSettings{
				ServerAddress: "testhost:1234",
				ServerHost:    "testhost",
				ServerPort:    1234,
				ServerDomain:  "testdomain",
				BaseURL:       "http://testhost:1234/",
			},
		},
		Flags: &config.Flags{
			ServerAddress: "testhost:1234",
			BaseURL:       "http://testhost:1234/",
		},
	}
}

func setupTestHandler(t *testing.T) (*gin.Engine, *mock.MockURLShortenerService, *config.Settings) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockURLShortenerService(ctrl)
	settings := getDefaultSettings()

	handler := NewCreatingShortLinksHandler(mockService, settings)

	// Создаем роутер с middleware для логгера
	logger, _ := zap.NewDevelopment()
	router := gin.New()
	router.Use(middleware.LoggerMiddleware(logger.Sugar()))

	// Добавляем мок JWT middleware для тестов
	router.Use(func(c *gin.Context) {
		// Создаем тестового пользователя
		testUser := &model.UserModel{
			ID:          "test_user_id",
			Name:        "test_user",
			Password:    "",
			IsAnonymous: false,
		}

		// Добавляем пользователя в контекст
		ctx := context.WithValue(c.Request.Context(), middleware.JWTTokenContextKey, testUser)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	})

	router.POST(apiPath, handler.Handle)

	return router, mockService, settings
}

func TestNewCreatingShortLinksHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockURLShortenerService(ctrl)
	settings := getDefaultSettings()

	handler := NewCreatingShortLinksHandler(mockService, settings)

	assert.NotNil(t, handler)
	assert.IsType(t, &CreatingShortLinks{}, handler)
}

func TestCreatingShortLinks_Handle_MissingContentType(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("https://test.com"))

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Отсутствует Content-Type")
}

func TestCreatingShortLinks_Handle_BodyReadError(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	// Передаем body, который вызывает ошибку при чтении
	badBody := &errReader{}
	req, _ := http.NewRequest(http.MethodPost, apiPath, badBody)
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка чтения данных")
}

type errReader struct{}

func (e *errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestCreatingShortLinks_Handle_EmptyBody(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("   "))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Пустой URL")
}

func TestCreatingShortLinks_Handle_WhitespaceURL(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("  \t\n  "))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Пустой URL")
}

func TestCreatingShortLinks_Handle_ServiceError(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://fail.com"
	expectedError := errors.New("service error")

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return("", expectedError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка при сокращении URL")
	assert.Contains(t, w.Body.String(), expectedError.Error())
}

func TestCreatingShortLinks_Handle_Success(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/very/long/url"
	expectedShortURL := "abc123"
	expectedResultURL := "http://testhost:1234/abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, expectedResultURL, w.Body.String())
}

func TestCreatingShortLinks_Handle_SuccessWithTrailingSlash(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "xyz789"
	expectedResultURL := "http://testhost:1234/xyz789"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, expectedResultURL, w.Body.String())
}

func TestCreatingShortLinks_Handle_ContextWithLogger(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")

	// Добавляем логгер в контекст
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())
	req = req.WithContext(ctx)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinks_Handle_RequestIDMiddleware(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Request-ID", "test-request-id")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinks_Handle_RemoteAddr(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")
	req.RemoteAddr = "192.168.1.1:12345"

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinks_Handle_ContentTypeVariations(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "text/plain",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "text/plain with charset",
			contentType:    "text/plain; charset=utf-8",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "application/json",
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty content type",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				mockService.EXPECT().
					ShortURL(gomock.Any(), gomock.Any()).
					Return("abc123", nil)
			} else {
				mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("https://test.com"))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreatingShortLinks_Handle_EdgeCases(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty body",
			requestBody:    "",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Пустой URL",
		},
		{
			name:           "only spaces",
			requestBody:    "   ",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Пустой URL",
		},
		{
			name:           "only tabs and newlines",
			requestBody:    "\t\n\r",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Пустой URL",
		},
		{
			name:           "normal URL",
			requestBody:    "https://example.com/test",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "URL with spaces around",
			requestBody:    "  https://example.com/test  ",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				// Настраиваем мок для успешного случая
				mockService.EXPECT().
					ShortURL(gomock.Any(), strings.TrimSpace(tt.requestBody)).
					Return("abc123", nil)
			} else {
				// Не ожидаем вызовов сервиса для ошибок
				mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestCreatingShortLinks_buildShortURL(t *testing.T) {
	handler := &CreatingShortLinks{
		baseURL: "http://testhost:1234/",
	}

	tests := []struct {
		name     string
		shortURL string
		expected string
	}{
		{
			name:     "normal case",
			shortURL: "abc123",
			expected: "http://testhost:1234/abc123",
		},
		{
			name:     "with trailing slash in baseURL",
			shortURL: "xyz789",
			expected: "http://testhost:1234/xyz789",
		},
		{
			name:     "empty short URL",
			shortURL: "",
			expected: "http://testhost:1234/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.buildShortURL(tt.shortURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreatingShortLinks_buildShortURL_WithoutTrailingSlash(t *testing.T) {
	handler := &CreatingShortLinks{
		baseURL: "http://testhost:1234", // без trailing slash
	}

	result := handler.buildShortURL("abc123")
	assert.Equal(t, "http://testhost:1234/abc123", result)
}

func TestCreatingShortLinks_Handle_ServiceErrorWithDetails(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://test.com"
	expectedError := fmt.Errorf("database connection failed")

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return("", expectedError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(longURL))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка при сокращении URL")
	assert.Contains(t, w.Body.String(), expectedError.Error())
}
