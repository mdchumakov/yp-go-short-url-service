package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для сервиса сокращения ссылок
type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) ShortenURL(longURL string) (string, error) {
	args := m.Called(longURL)
	return args.String(0), args.Error(1)
}

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
			SQLite: &db.SQLiteSettings{
				SQLiteDBPath: "test.db",
			},
		},
		Flags: &config.Flags{
			ServerAddress: "testhost:1234",
			BaseURL:       "http://testhost:1234/",
		},
	}
}

func TestCreatingShortLinks_Handle_NoContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksHandler(mockService, getDefaultSettings())
	r.POST("/shorten", handler.Handle)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("https://test.com"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Отсутствует Content-Type")
}

func TestCreatingShortLinks_Handle_BodyReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksHandler(mockService, getDefaultSettings())
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

func (e *errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestCreatingShortLinks_Handle_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksHandler(mockService, getDefaultSettings())
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
	handler := NewCreatingShortLinksHandler(mockService, getDefaultSettings())
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

	expectedShortURL := "abc123"
	settings := getDefaultSettings()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksHandler(mockService, settings)
	r.POST("/shorten", handler.Handle)

	mockService.On("ShortenURL", "https://ok.com").Return(expectedShortURL, nil)

	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader("https://ok.com"))
	req.Header.Set("Content-Type", "text/plain")

	r.ServeHTTP(w, req)

	expectedURL := fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(handler.baseURL, "/"),
		expectedShortURL,
	)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), expectedURL)
}
