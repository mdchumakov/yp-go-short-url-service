package handler

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.30.1"))
	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestExtractingFullLink_Handle_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	db, mock := setupTestDB(t)

	shortURL := "1a2b3c"
	longURL := "https://example.com/very/long/url"

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "short_url", "long_url"}).
		AddRow(1, time.Now(), time.Now(), shortURL, longURL)

	expectedSQL := regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")
	mock.ExpectQuery(expectedSQL).WithArgs(shortURL).WillReturnRows(rows)

	// Настройка Gin
	w := httptest.NewRecorder()
	r := gin.New()
	handler := NewExtractingFullLink(db)
	r.GET("/:shortURL", handler.Handle)

	req, _ := http.NewRequest(http.MethodGet, "/"+shortURL, nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, longURL, w.Header().Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExtractingFullLink_Handle_NotFound(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	db, mock := setupTestDB(t)

	shortURL := "notfound"

	expectedSQL := regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")
	mock.ExpectQuery(expectedSQL).WithArgs(shortURL).WillReturnError(gorm.ErrRecordNotFound)

	// Настройка Gin
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	handler := NewExtractingFullLink(db)
	r.GET("/:shortURL", handler.Handle)

	// Мы должны установить параметр URL в тестовом контексте
	c.Request, _ = http.NewRequest(http.MethodGet, "/"+shortURL, nil)
	c.Params = gin.Params{gin.Param{Key: "shortURL", Value: shortURL}}

	// Act
	r.ServeHTTP(w, c.Request)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Ссылка не найдена", w.Body.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}
