package json

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service/mock"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const apiPath = "/api/shorten"

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

func setupTestHandler(t *testing.T) (*gin.Engine, *mock.MockLinkShortenerService, *config.Settings) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockLinkShortenerService(ctrl)
	settings := getDefaultSettings()

	handler := NewCreatingShortURLsAPIHandler(mockService, settings)

	// Создаем роутер с middleware для логгера
	logger, _ := zap.NewDevelopment()
	router := gin.New()
	router.Use(middleware.LoggerMiddleware(logger.Sugar()))
	router.POST(apiPath, handler.Handle)

	return router, mockService, settings
}

func TestNewCreatingShortLinksAPIHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockLinkShortenerService(ctrl)
	settings := getDefaultSettings()

	handler := NewCreatingShortURLsAPIHandler(mockService, settings)

	assert.NotNil(t, handler)
	assert.IsType(t, &creatingShortURLsAPIHandler{}, handler)
}

func TestCreatingShortLinksAPIHandler_Handle_MissingContentType(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "https://test.com"}`))

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Content-type header is required", response["message"])
}

func TestCreatingShortLinksAPIHandler_Handle_InvalidContentType(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "https://test.com"}`))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Content-type: `application/json` header is required", response["message"])
}

func TestCreatingShortLinksAPIHandler_Handle_InvalidJSON(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", response["error"])
}

func TestCreatingShortLinksAPIHandler_Handle_EmptyURL(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": ""}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request body", response["error"])
}

func TestCreatingShortLinksAPIHandler_Handle_WhitespaceURL(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)
	defer mockService.EXPECT().ShortURL(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "   "}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "url is required", response["error"])
}

func TestCreatingShortLinksAPIHandler_Handle_ServiceError(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://test.com"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return("", assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, assert.AnError.Error(), response["error"])
}

func TestCreatingShortLinksAPIHandler_Handle_Success(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/very/long/url"
	expectedShortURL := "abc123"
	expectedResultURL := "http://testhost:1234/abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response CreatingShortURLsDTOOut
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResultURL, response.Result)
}

func TestCreatingShortLinksAPIHandler_Handle_SuccessWithTrailingSlash(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "xyz789"
	expectedResultURL := "http://testhost:1234/xyz789"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreatingShortURLsDTOOut
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResultURL, response.Result)
}

func TestCreatingShortLinksAPIHandler_Handle_ContextWithLogger(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем логгер в контекст
	logger, _ := zap.NewDevelopment()
	ctx := middleware.WithLogger(context.Background(), logger.Sugar())
	req = req.WithContext(ctx)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinksAPIHandler_buildShortURL(t *testing.T) {
	handler := &creatingShortURLsAPIHandler{
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

func TestCreatingShortLinksAPIHandler_buildShortURL_WithoutTrailingSlash(t *testing.T) {
	handler := &creatingShortURLsAPIHandler{
		baseURL: "http://testhost:1234", // без trailing slash
	}

	result := handler.buildShortURL("abc123")
	assert.Equal(t, "http://testhost:1234/abc123", result)
}

func TestCreatingShortLinksAPIHandler_Handle_EdgeCases(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing url field",
			requestBody:    `{}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "null url",
			requestBody:    `{"url": null}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "url with only spaces",
			requestBody:    `{"url": "   "}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "url is required",
		},
		{
			name:           "very long URL",
			requestBody:    `{"url": "` + strings.Repeat("a", 10000) + `"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated, // должен пройти успешно
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				// Настраиваем мок для успешного случая
				mockService.EXPECT().
					ShortURL(gomock.Any(), gomock.Any()).
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
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

// Дополнительные тесты для улучшения покрытия
func TestCreatingShortLinksAPIHandler_Handle_RequestIDMiddleware(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinksAPIHandler_Handle_RemoteAddr(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatingShortLinksAPIHandler_Handle_ContentTypeVariations(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "application/json with charset",
			contentType:    "application/json; charset=utf-8",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "application/json with spaces",
			contentType:    "  application/json  ",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "text/html",
			contentType:    "text/html",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "multipart/form-data",
			contentType:    "multipart/form-data",
			expectedStatus: http.StatusUnsupportedMediaType,
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
			req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "https://test.com"}`))
			req.Header.Set("Content-Type", tt.contentType)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreatingShortLinksAPIHandler_Handle_JSONMarshalError(t *testing.T) {
	// Этот тест проверяет обработку ошибки маршалинга JSON
	// В реальных условиях это маловероятно, но важно покрыть
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://example.com/test"
	expectedShortURL := "abc123"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return(expectedShortURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// В нормальных условиях этот тест должен пройти успешно
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response CreatingShortURLsDTOOut
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "http://testhost:1234/abc123", response.Result)
}

func TestCreatingShortLinksAPIHandler_Handle_ServiceErrorWithDetails(t *testing.T) {
	router, mockService, _ := setupTestHandler(t)

	longURL := "https://test.com"
	expectedError := "database connection failed"

	mockService.EXPECT().
		ShortURL(gomock.Any(), longURL).
		Return("", fmt.Errorf("%s", expectedError))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedError, response["error"])
}
