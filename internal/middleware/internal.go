package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type InternalMiddleware struct {
	logger        *zap.SugaredLogger
	trustedSubnet string
}

func NewInternalMiddleware(logger *zap.SugaredLogger, trustedSubnet string) *InternalMiddleware {
	return &InternalMiddleware{
		logger:        logger,
		trustedSubnet: trustedSubnet,
	}
}

func (m *InternalMiddleware) InternalMiddlewareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := ExtractRequestID(c.Request.Context())

		m.logger.Infow("handling request", "id", requestID)
		if !m.isTrustedSubnet(c) {
			c.String(http.StatusForbidden, "trusted subnet is not allowed")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *InternalMiddleware) isTrustedSubnet(c *gin.Context) bool {

	requestID := ExtractRequestID(c.Request.Context())

	m.logger.Infow("checking subnet", "id", requestID)

	// Получаем IP-адрес из заголовка X-Real-IP
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP == "" {
		m.logger.Warnw("X-Real-IP header is missing", "id", requestID)
		return false
	}

	// При пустом значении переменной trusted_subnet
	// доступ к эндпоинту должен быть запрещён для любого входящего запроса.
	if m.trustedSubnet == "" {
		m.logger.Warnw("trusted subnet is not configured", "id", requestID)
		return false
	}

	// Парсим IP-адрес клиента
	clientIP := net.ParseIP(xRealIP)
	if clientIP == nil {
		m.logger.Warnw(
			"invalid IP address in X-Real-IP header",
			"ip", xRealIP,
			"id", requestID,
		)
		return false
	}

	// Парсим CIDR подсеть
	_, ipNet, err := net.ParseCIDR(m.trustedSubnet)
	if err != nil {
		m.logger.Errorw(
			"invalid trusted subnet CIDR",
			"subnet", m.trustedSubnet,
			"error", err,
			"id", requestID,
		)
		return false
	}

	// Проверяем, входит ли IP-адрес в доверенную подсеть
	isInSubnet := ipNet.Contains(clientIP)

	m.logger.Infow("subnet check result",
		"ip", xRealIP,
		"subnet", m.trustedSubnet,
		"allowed", isInSubnet,
		"id", requestID)

	return isInSubnet
}
