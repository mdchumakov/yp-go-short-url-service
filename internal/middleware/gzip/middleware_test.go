package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	shortenAPI "yp-go-short-url-service/internal/handler/api/shorten"
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

func TestCreatingShortLinksAPI_Handle_GZIPRequest_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	logger, _ := config.NewDevLogger()
	defer config.SyncLogger(logger)

	expectedShortURL := "abc123"
	settings := getDefaultSettings()

	mockService := new(MockShortener)
	handler := shortenAPI.NewCreatingShortLinksAPI(mockService, settings)

	r.Use(Middleware(logger))

	r.POST(apiPath, handler.Handle)

	longURL := "https://ok.com"
	mockService.On("ShortenURL", longURL).Return(expectedShortURL, nil)

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte(`{"url": "` + longURL + `"}`))
	_ = zw.Close()

	req, _ := http.NewRequest(
		http.MethodPost,
		apiPath,
		&buf,
	)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	expectedBody := shortenAPI.CreatingShortLinksDTOOut{Result: settings.GetBaseURL() + expectedShortURL}
	expectedBodyJSON, _ := json.Marshal(expectedBody)

	zr, _ := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	data, err := io.ReadAll(zr)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, string(expectedBodyJSON), string(data))
	mockService.AssertExpectations(t)
}
