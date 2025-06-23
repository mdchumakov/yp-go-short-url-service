package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yp-go-short-url-service/internal/config"
)

// Мок для сервиса сокращения ссылок
type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) ShortenURL(longURL string) (string, error) {
	args := m.Called(longURL)
	return args.String(0), args.Error(1)
}

func getDefaultSettings() *config.ServerSettings {
	return &config.ServerSettings{
		ServerHost:   "localhost",
		ServerPort:   8080,
		ServerDomain: "localhost",
	}
}

func TestCreatingShortLinks_Handle_NoContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinks(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("https://test.com"))
	// Не устанавливаем Content-Type

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Отсутствует Content-Type")
}

func TestCreatingShortLinks_Handle_BodyReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinks(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	// Передаем body, который вызывает ошибку при чтении
	badBody := &errReader{}
	req, _ := http.NewRequest(http.MethodPost, "/shorten", badBody)
	req.Header.Set("Content-Type", "text/plain")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка чтения данных")
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestCreatingShortLinks_Handle_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinks(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("   "))
	req.Header.Set("Content-Type", "text/plain")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Пустой URL")
}

func TestCreatingShortLinks_Handle_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinks(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	mockService.On("ShortenURL", "https://fail.com").Return("", errors.New("fail"))

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("https://fail.com"))
	req.Header.Set("Content-Type", "text/plain")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка при сокращении URL")
}

func TestCreatingShortLinks_Handle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinks(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	mockService.On("ShortenURL", "https://ok.com").Return("abc123", nil)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("https://ok.com"))
	req.Header.Set("Content-Type", "text/plain")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "http://localhost:8080/abc123")
}
