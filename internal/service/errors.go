package service

import "errors"

var (
	// ErrURLWasDeleted возвращается, когда запрашиваемый URL был удален.
	ErrURLWasDeleted = errors.New("url was deleted")
	// ErrURLAlreadyExists возвращается, когда пытаются создать короткий URL для уже существующего длинного URL.
	ErrURLAlreadyExists = errors.New("url already exists")
)

// IsAlreadyExistsError проверяет, является ли ошибка ошибкой "URL уже существует".
// Возвращает true, если ошибка равна ErrURLAlreadyExists.
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrURLAlreadyExists)
}

// IsDeletedError проверяет, является ли ошибка ошибкой "URL был удален".
// Возвращает true, если ошибка равна ErrURLWasDeleted.
func IsDeletedError(err error) bool {
	return errors.Is(err, ErrURLWasDeleted)
}
