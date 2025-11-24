package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Middleware(log *zap.SugaredLogger) gin.HandlerFunc {
	// Поддержка gzip
	return func(c *gin.Context) {

		// Функция сжатия должна работать для контента с типами application/json и text/html.
		// А также, что удивительно, для application/x-gzip. (В задании про него не сказано, но в CI он есть)
		if contentType := c.Request.Header.Get("Content-Type"); !checkContent(contentType) {
			log.Debugw(
				"skipping gzip middleware for unsupported content type",
				"contentType",
				contentType,
			)
			c.Next()
			return
		}
		log.Debugw("applying gzip middleware for",
			"contentType", c.Request.Header.Get("Content-Type"),
			"contentEncoding", c.Request.Header.Get("Content-Encoding"),
		)

		// Сервис должен уметь принимать запросы в сжатом формате (с HTTP-заголовком Content-Encoding).
		if err := decompressRequest(c, log); err != nil {
			log.Errorw("failed to decompress request", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем, нужно ли сжимать ответ
		if contentEncoding := c.Request.Header.Get("Accept-Encoding"); !strings.Contains(contentEncoding, "gzip") {
			log.Warnw("skipping gzip middleware for unsupported content Encoding",
				"contentEncoding", contentEncoding,
			)
			c.Next()
			return
		}

		// Сервис должен уметь возвращать ответ в сжатом формате (с HTTP-заголовком Content-Encoding).
		gw := gzip.NewWriter(c.Writer)

		// Создаем кастомный response writer
		writer := &gzipResponseBodyWriter{
			ResponseWriter: c.Writer,
			writer:         gw,
		}
		c.Writer = writer

		// Устанавливаем заголовки для сжатого ответа
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", c.Request.Header.Get("Content-Type"))
		c.Writer.Header().Add("Vary", "Accept-Encoding")

		c.Next()

		// Закрываем gzip writer через response writer
		if err := writer.Close(); err != nil {
			log.Errorw("failed to close gzip writer", "error", err)
		}
	}
}

func checkContent(contentType string) bool {
	// Проверяем, что Content-Type соответствует одному из поддерживаемых типов
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/x-gzip") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "plain/text") ||
		strings.Contains(contentType, "text/plain")
}

func decompressRequest(c *gin.Context, log *zap.SugaredLogger) error {
	if c.Request.Header.Get("Content-Encoding") != "gzip" {
		return nil
	}

	gz, err := gzip.NewReader(c.Request.Body)
	if err != nil {
		return err
	}

	defer func() {
		if err := gz.Close(); err != nil {
			log.Errorw("failed to close gzip writer", "error", err)
		}
	}()

	c.Request.Body = gz
	return nil
}
