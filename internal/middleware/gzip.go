package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type gzipResponseBodyWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (r *gzipResponseBodyWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *gzipResponseBodyWriter) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func GZIPMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	// Поддержка gzip
	return func(c *gin.Context) {

		// Функция сжатия должна работать для контента с типами application/json и text/html.
		if contentType := c.Request.Header.Get("Content-Type"); contentType != "application/json" && contentType != "text/html" {
			log.Debugw("skipping gzip middleware for unsupported content type", "contentType", contentType)
			c.Next()
			return
		}
		// Сервис должен уметь принимать запросы в сжатом формате (с HTTP-заголовком Content-Encoding).
		if err := decompressRequest(c, log); err != nil {
			log.Errorw("failed to decompress request", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Подменяем writer
		gz := gzip.NewWriter(c.Writer)
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				log.Errorw("failed to close gzip writer", "error", err)
			}
		}(gz)
		writer := &gzipResponseBodyWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}
		c.Writer = writer

		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", c.Request.Header.Get("Content-Type"))
		c.Writer.Header().Add("Vary", "Accept-Encoding")

		c.Next()
	}
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
