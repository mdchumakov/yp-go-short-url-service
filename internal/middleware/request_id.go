package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"
const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware добавляет уникальный ID к каждому запросу для трассировки
func RequestIDMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, есть ли уже Request ID в заголовке
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем Request ID в контекст Gin
		c.Set(RequestIDKey, requestID)

		// Добавляем Request ID в заголовки ответа для клиента
		c.Header(RequestIDHeader, requestID)

		logger.Debugw("Request ID generated",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path)

		c.Next()
	}
}

func ExtractRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}
