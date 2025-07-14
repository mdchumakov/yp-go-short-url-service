package service

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestShortenURL_ExistingURL(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	service := NewLinkShortenerService(db)

	longURL := "https://existing-url.com"
	existingShortURL := "9AX8jYU3"

	rows := sqlmock.NewRows([]string{"id", "short_url"}).AddRow(1, existingShortURL)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE long_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(longURL).
		WillReturnRows(rows)

	// Act
	shortURL, err := service.ShortenURL(longURL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingShortURL, shortURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenURL_NewURL_NoConflict(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	service := NewLinkShortenerService(db)

	longURL := "https://new-url.com"
	generatedShortURL := shortenURLBase62(longURL) // "fofX6g"

	// 1. Проверка на существование longURL
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE long_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(longURL).
		WillReturnError(gorm.ErrRecordNotFound)

	// 2. Проверка на конфликт с shortURL
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(generatedShortURL).
		WillReturnError(gorm.ErrRecordNotFound)

	// 3. Создание новой записи
	mock.ExpectBegin()
	mock.ExpectExec(
		regexp.QuoteMeta("INSERT INTO `urls` (`short_url`,`long_url`,`created_at`,`updated_at`) VALUES (?,?,?,?)")).
		WithArgs(generatedShortURL, longURL, AnyTime{}, AnyTime{}).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	shortURL, err := service.ShortenURL(longURL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, generatedShortURL, shortURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenURL_NewURL_WithConflict(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	service := NewLinkShortenerService(db)

	longURL := "https://colliding-url.com"
	initialShortURL := shortenURLBase62(longURL)
	collidingURLPlus1 := initialShortURL + "1"
	resolvedShortURL := shortenURLBase62(collidingURLPlus1)

	// 1. Проверка longURL -> не найден
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE long_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(longURL).
		WillReturnError(gorm.ErrRecordNotFound)

	// 2. Проверка initialShortURL -> найден (конфликт!)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(initialShortURL).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// 3. resolveConflicts: проверка initialShortURL еще раз -> найден
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(initialShortURL).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// 4. resolveConflicts: проверка resolvedShortURL -> не найден
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE short_url = ? ORDER BY `urls`.`id` LIMIT 1")).
		WithArgs(resolvedShortURL).
		WillReturnError(gorm.ErrRecordNotFound)

	// 5. Создание записи с resolvedShortURL
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `urls` (`short_url`,`long_url`,`created_at`,`updated_at`) VALUES (?,?,?,?)")).
		WithArgs(resolvedShortURL, longURL, AnyTime{}, AnyTime{}).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	// Act
	shortURL, err := service.ShortenURL(longURL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, resolvedShortURL, shortURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}
