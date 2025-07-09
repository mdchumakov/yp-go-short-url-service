package shorten

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
)

const apiPath = "/api/shorten"

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

func TestCreatingShortLinksAPI_Handle_NoContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksAPI(mockService, getDefaultSettings())
	r.POST(apiPath, handler.Handle)

	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("https://test.com"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"message":"Content-type header is required"}`, w.Body.String())
}

func TestCreatingShortLinksAPI_Handle_BodyReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksAPI(mockService, getDefaultSettings())
	r.POST(apiPath, handler.Handle)

	// Передаем body, который вызывает ошибку при чтении
	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "invalid request body"}`, w.Body.String())
}

func TestCreatingShortLinksAPI_Handle_EmptyURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksAPI(mockService, getDefaultSettings())
	r.POST(apiPath, handler.Handle)

	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": ""}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "url is required"}`, w.Body.String())
}

func TestCreatingShortLinksAPI_Handle_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksAPI(mockService, getDefaultSettings())
	r.POST(apiPath, handler.Handle)

	longURL := "https://test.com"
	mockService.On("ShortenURL", longURL).Return("", assert.AnError)

	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "`+assert.AnError.Error()+`"}`, w.Body.String())
	mockService.AssertExpectations(t)
}

func TestCreatingShortLinksAPI_Handle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	expectedShortURL := "abc123"
	settings := getDefaultSettings()

	mockService := new(MockShortener)
	handler := NewCreatingShortLinksAPI(mockService, settings)
	r.POST(apiPath, handler.Handle)

	longURL := "https://ok.com"
	mockService.On("ShortenURL", longURL).Return(expectedShortURL, nil)

	req, _ := http.NewRequest(http.MethodPost, apiPath, strings.NewReader(`{"url": "`+longURL+`"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	resultURL := settings.GetBaseURL() + expectedShortURL
	assert.Equal(t, `{"result":"`+resultURL+`"}`, w.Body.String())
	mockService.AssertExpectations(t)
}
