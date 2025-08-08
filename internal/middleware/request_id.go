package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKeyType - пользовательский тип для ключа Request ID в контексте
type RequestIDKeyType struct{}

// RequestIDKey - ключ для хранения Request ID в контексте
var RequestIDKey = RequestIDKeyType{}

const RequestIDHeader = "X-Request-ID"
const GinRequestIDKey = "request_id"

// RequestIDMiddleware добавляет уникальный ID к каждому запросу для трассировки
func RequestIDMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, есть ли уже Request ID в заголовке
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем Request ID в контекст Gin
		c.Set(GinRequestIDKey, requestID)

		// Добавляем Request ID в обычный контекст для функции ExtractRequestID
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

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
