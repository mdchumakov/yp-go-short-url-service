package destructor

import (
	"errors"
	"strings"
)

const base62Chars string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// DestructorRequestBodyDTOIn представляет входные данные для удаления URL пользователя
// Массив коротких URL для удаления
// required: true
// example: ["6qxTVvsy", "RTfd56hn", "Jlfd67ds"]
type DestructorRequestBodyDTOIn []string

// Validate проверяет валидность DTO и очищает данные
func (dto DestructorRequestBodyDTOIn) Validate() error {
	if len(dto) == 0 {
		return errors.New("urls array cannot be empty")
	}

	for i, shortURL := range dto {
		// Убираем пробелы
		shortURL = strings.TrimSpace(shortURL)

		if shortURL == "" {
			return errors.New("short URL cannot be empty")
		}

		if len(shortURL) < 3 {
			return errors.New("short URL is too short")
		}

		if len(shortURL) > 50 {
			return errors.New("short URL is too long")
		}

		// Проверяем, что URL содержит только допустимые символы
		for _, char := range shortURL {
			if !strings.ContainsRune(base62Chars, char) {
				return errors.New("short URL contains invalid characters")
			}
		}

		// Обновляем значение в массиве (убираем пробелы)
		dto[i] = shortURL
	}

	return nil
}
