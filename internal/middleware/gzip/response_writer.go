package gzip

import (
	"compress/gzip"

	"github.com/gin-gonic/gin"
)

// gzipResponseBodyWriter обертка над gin.ResponseWriter для записи сжатых ответов в формате gzip.
// Перехватывает запись данных и сжимает их перед отправкой клиенту.
type gzipResponseBodyWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
	closed bool
}

// Write записывает данные в gzip writer для сжатия перед отправкой клиенту.
// Возвращает количество записанных байт и ошибку, если запись не удалась.
func (r *gzipResponseBodyWriter) Write(b []byte) (int, error) {
	if r.closed {
		return 0, nil
	}
	return r.writer.Write(b)
}

// WriteString записывает строку в gzip writer для сжатия перед отправкой клиенту.
// Возвращает количество записанных байт и ошибку, если запись не удалась.
func (r *gzipResponseBodyWriter) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

// Close закрывает gzip writer и завершает сжатие данных.
// Возвращает ошибку, если закрытие не удалось. Безопасно вызывать несколько раз.
func (r *gzipResponseBodyWriter) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.writer.Close()
}
