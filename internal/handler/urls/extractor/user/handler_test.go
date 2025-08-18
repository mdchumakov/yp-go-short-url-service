package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service/mock"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestHandler(t *testing.T) (*gin.Engine, *mock.MockLinkExtractorService) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockLinkExtractorService(ctrl)

	handler := NewExtractingUserURLsHandler(mockService)

	logger, _ := zap.NewDevelopment()
	router := gin.New()
	router.Use(middleware.LoggerMiddleware(logger.Sugar()))
	router.Use(middleware.RequestIDMiddleware(logger.Sugar()))
	router.GET("/api/user/urls", handler.Handle)

	return router, mockService
}

func TestExtractingUserURLsHandler_Handle_Success(t *testing.T) {
	router, mockService := setupTestHandler(t)

	// Подготавливаем тестовые данные
	testURLs := []*model.URLsModel{
		{
			ID:        1,
			ShortURL:  "abc123",
			LongURL:   "https://example.com/long-url-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			ShortURL:  "def456",
			LongURL:   "https://example.com/long-url-2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Настраиваем мок
	mockService.EXPECT().
		ExtractUserURLs(gomock.Any(), "test-user-id").
		Return(testURLs, nil)

	// Создаем запрос
	req, _ := http.NewRequest("GET", "/api/user/urls", nil)

	// Добавляем пользователя в контекст
	ctx := context.WithValue(req.Context(), middleware.JWTTokenContextKey, &model.UserModel{
		ID: "test-user-id",
	})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)

	var response UserURLsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "abc123", response[0].ShortURL)
	assert.Equal(t, "https://example.com/long-url-1", response[0].OriginalURL)
	assert.Equal(t, "def456", response[1].ShortURL)
	assert.Equal(t, "https://example.com/long-url-2", response[1].OriginalURL)
}

func TestExtractingUserURLsHandler_Handle_NoUser(t *testing.T) {
	router, _ := setupTestHandler(t)

	// Создаем запрос без пользователя в контексте
	req, _ := http.NewRequest("GET", "/api/user/urls", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
