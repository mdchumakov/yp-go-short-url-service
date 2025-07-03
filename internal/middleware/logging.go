package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the request details
		log.Infow("Request received",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"remote_addr", c.Request.RemoteAddr,
		)

		// Process the request
		c.Next()

		// Log the response status
		log.Infow("Response sent",
			"status", c.Writer.Status(),
			"path", c.Request.URL.Path,
		)
	}
}
