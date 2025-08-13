package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerKey - ключ для хранения логгера в контексте
type loggerKey struct{}

// WithLogger добавляет логгер в контекст
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger извлекает логгер из контекста
func GetLogger(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return logger
	}
	// Возвращаем no-op логгер если логгер не найден
	return zap.NewNop().Sugar()
}

// LoggerMiddleware добавляет логгер в контекст запроса
func LoggerMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Добавляем логгер в контекст
		ctx := WithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		timeStart := time.Now()

		// Сведения о запросах должны содержать URL, метод запроса.
		log.Infow("Request received",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"remoteAddr", c.Request.RemoteAddr,
		)

		c.Next()

		duration := time.Since(timeStart)

		// Сведения об ответах должны содержать код статуса,
		// размер содержимого ответа и время, затраченное на его выполнение.
		log.Infow("Response sent",
			"status", c.Writer.Status(),
			"path", c.Request.URL.Path,
			"duration", duration,
			"byteSize", c.Writer.Size(),
		)
	}
}
