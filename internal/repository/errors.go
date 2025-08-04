package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var (
	ErrURLNotFound = errors.New("URL не найден")
	ErrURLExists   = errors.New("URL уже существует")
)

// IsNotFoundError проверяет, является ли ошибка ошибкой "не найдено"
func IsNotFoundError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrURLNotFound)
}

// IsExistsError проверяет, является ли ошибка ошибкой "уже существует"
func IsExistsError(err error) bool {
	return errors.Is(err, ErrURLExists)
}
