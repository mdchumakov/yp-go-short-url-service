package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Ошибки репозитория, используемые для обработки различных ситуаций при работе с базой данных.
var (
	// ErrURLNotFound возвращается, когда запрашиваемый URL не найден в базе данных.
	ErrURLNotFound = errors.New("URL не найден")
	// ErrURLExists возвращается, когда пытаются создать URL, который уже существует в базе данных.
	ErrURLExists = errors.New("URL уже существует")
	// ErrUserNotFound возвращается, когда запрашиваемый пользователь не найден в базе данных.
	ErrUserNotFound = errors.New("пользователь не найден")
)

// IsNotFoundError проверяет, является ли ошибка ошибкой "не найдено"
func IsNotFoundError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrURLNotFound)
}

// IsExistsError проверяет, является ли ошибка ошибкой "уже существует"
func IsExistsError(err error) bool {
	if errors.Is(err, ErrURLExists) {
		return true
	}

	// Проверяем PostgreSQL ошибки
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL код для unique_violation
		if pgErr.Code == "23505" {
			return true
		}
	}

	// Проверяем SQLite ошибки (если используется database/sql)
	// SQLite возвращает "UNIQUE constraint failed" в сообщении
	if err != nil && (err.Error() == "UNIQUE constraint failed" ||
		err.Error() == "UNIQUE constraint failed: urls.short_url" ||
		err.Error() == "UNIQUE constraint failed: urls.long_url") {
		return true
	}

	return false
}
