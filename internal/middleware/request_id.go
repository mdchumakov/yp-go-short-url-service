package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKeyType - пользовательский тип для ключа Request ID в контексте
type RequestIDKeyType struct{}

// RequestIDKey - ключ для хранения Request ID в контексте.
// Используется для извлечения уникального идентификатора запроса из контекста.
var RequestIDKey = RequestIDKeyType{}

// RequestIDHeader - название HTTP заголовка для передачи уникального идентификатора запроса.
const RequestIDHeader = "X-Request-ID"

// GinRequestIDKey - ключ для хранения Request ID в контексте Gin.
const GinRequestIDKey = "request_id"

// RequestIDMiddleware добавляет уникальный ID к каждому запросу для трассировки.
// Генерирует новый UUID, если Request ID отсутствует в заголовке X-Request-ID.
// Добавляет Request ID в контекст запроса и в заголовки ответа.
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

// ExtractRequestID извлекает Request ID из контекста запроса.
// Возвращает "unknown", если Request ID не найден в контексте.
func ExtractRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}
