package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func LoggerMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
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
