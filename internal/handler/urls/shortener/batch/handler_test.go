package batch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/service/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreatingShortURLsByBatchAPIHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		baseURL        string
		mockSetup      func(*mock.MockURLShortenerService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное сокращение URL",
			requestBody: `[
				{
					"correlation_id": "1",
					"original_url": "https://example.com/very-long-url-1"
				},
				{
					"correlation_id": "2",
					"original_url": "https://example.com/very-long-url-2"
				}
			]`,
			contentType: "application/json",
			baseURL:     "http://localhost:8080",
			mockSetup: func(mockService *mock.MockURLShortenerService) {
				expectedInput := []map[string]string{
					{"correlation_id": "1", "original_url": "https://example.com/very-long-url-1"},
					{"correlation_id": "2", "original_url": "https://example.com/very-long-url-2"},
				}
				mockService.EXPECT().
					ShortURLsByBatch(gomock.Any(), expectedInput).
					Return([]map[string]string{
						{"correlation_id": "1", "short_url": "abc123"},
						{"correlation_id": "2", "short_url": "def456"},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: `[
				{
					"correlation_id": "1",
					"short_url": "http://localhost:8080/abc123"
				},
				{
					"correlation_id": "2",
					"short_url": "http://localhost:8080/def456"
				}
			]`,
		},
		{
			name:           "отсутствует Content-Type",
			requestBody:    `[]`,
			contentType:    "",
			baseURL:        "http://localhost:8080",
			mockSetup:      func(mockService *mock.MockURLShortenerService) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   `{"message":"Content-type header is required"}`,
		},
		{
			name:           "неверный Content-Type",
			requestBody:    `[]`,
			contentType:    "text/plain",
			baseURL:        "http://localhost:8080",
			mockSetup:      func(mockService *mock.MockURLShortenerService) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   `{"message":"Content-type: ` + "`" + `application/json` + "`" + ` header is required"}`,
		},
		{
			name:           "неверный JSON",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}`,
			contentType:    "application/json",
			baseURL:        "http://localhost:8080",
			mockSetup:      func(mockService *mock.MockURLShortenerService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unexpected EOF"}`,
		},
		{
			name:           "пустой массив URL",
			requestBody:    `[]`,
			contentType:    "application/json",
			baseURL:        "http://localhost:8080",
			mockSetup:      func(mockService *mock.MockURLShortenerService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"urls array is empty"}`,
		},
		{
			name: "ошибка сервиса",
			requestBody: `[
				{
					"correlation_id": "1",
					"original_url": "https://example.com/very-long-url-1"
				}
			]`,
			contentType: "application/json",
			baseURL:     "http://localhost:8080",
			mockSetup: func(mockService *mock.MockURLShortenerService) {
				expectedInput := []map[string]string{
					{"correlation_id": "1", "original_url": "https://example.com/very-long-url-1"},
				}
				mockService.EXPECT().
					ShortURLsByBatch(gomock.Any(), expectedInput).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock.NewMockURLShortenerService(ctrl)
			tt.mockSetup(mockService)

			// Создаем мок настроек для тестов
			settings := &config.Settings{
				EnvSettings: &config.ENVSettings{
					Server: &config.ServerSettings{
						BaseURL: tt.baseURL,
					},
				},
			}

			handler := NewCreatingShortURLsByBatchAPIHandler(mockService, settings)

			// Создаем HTTP запрос
			req, err := http.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Создаем HTTP recorder
			w := httptest.NewRecorder()

			// Создаем Gin контекст
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/api/shorten/batch", handler.Handle)

			// Выполняем запрос
			router.ServeHTTP(w, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем тело ответа
			if tt.expectedBody != "" {
				// Для JSON ответов нормализуем форматирование
				if tt.expectedStatus == http.StatusCreated {
					var expected, actual interface{}
					err := json.Unmarshal([]byte(tt.expectedBody), &expected)
					require.NoError(t, err)
					err = json.Unmarshal(w.Body.Bytes(), &actual)
					require.NoError(t, err)
					assert.Equal(t, expected, actual)
				} else {
					assert.JSONEq(t, tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestCreatingShortURLsByBatchAPIHandler_convertToDTOOut(t *testing.T) {
	handler := &creatingShortURLsByBatchAPIHandler{}

	input := []map[string]string{
		{"correlation_id": "1", "short_url": "http://localhost:8080/abc123"},
		{"correlation_id": "2", "short_url": "http://localhost:8080/def456"},
	}

	expected := CreatingShortURLsByBatchDTOOut{
		{CorrelationID: "1", ShortURL: "http://localhost:8080/abc123"},
		{CorrelationID: "2", ShortURL: "http://localhost:8080/def456"},
	}

	result := handler.convertToDTOOut(input)
	assert.Equal(t, expected, result)
}

func TestCreatingShortURLsByBatchAPIHandler_buildShortURL(t *testing.T) {
	handler := &creatingShortURLsByBatchAPIHandler{
		baseURL: "http://localhost:8080",
	}

	input := []map[string]string{
		{"correlation_id": "1", "short_url": "abc123"},
		{"correlation_id": "2", "short_url": "def456"},
	}

	expected := []map[string]string{
		{"correlation_id": "1", "short_url": "http://localhost:8080/abc123"},
		{"correlation_id": "2", "short_url": "http://localhost:8080/def456"},
	}

	err := handler.buildShortURL(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, input)
}

func TestCreatingShortURLsByBatchAPIHandler_buildShortURL_WithTrailingSlash(t *testing.T) {
	handler := &creatingShortURLsByBatchAPIHandler{
		baseURL: "http://localhost:8080/",
	}

	input := []map[string]string{
		{"correlation_id": "1", "short_url": "abc123"},
	}

	expected := []map[string]string{
		{"correlation_id": "1", "short_url": "http://localhost:8080/abc123"},
	}

	err := handler.buildShortURL(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, input)
}

func TestCreatingShortURLsByBatchDTOIn_ToMapSlice(t *testing.T) {
	dto := CreatingShortURLsByBatchDTOIn{
		{CorrelationID: "1", OriginalURL: "https://example.com/1"},
		{CorrelationID: "2", OriginalURL: "https://example.com/2"},
	}

	expected := []map[string]string{
		{"correlation_id": "1", "original_url": "https://example.com/1"},
		{"correlation_id": "2", "original_url": "https://example.com/2"},
	}

	result := dto.ToMapSlice()
	assert.Equal(t, expected, result)
}
