package stats

import (
	"net"
	"net/http"
	handler_ "yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

type handler struct {
	trustedSubnet string
	service       service.StatsService
}

func New(service service.StatsService, trustedSubnet string) handler_.Handler {
	return &handler{
		service:       service,
		trustedSubnet: trustedSubnet,
	}
}

func (h *handler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	logger.Infow("handling request", "id", requestID)
	if !h.isTrustedSubnet(c) {
		c.String(http.StatusForbidden, "trusted subnet is not allowed")
	}

	usersCount, err := h.service.GetTotalUsersCount(c.Request.Context())
	if err != nil {
		logger.Errorw("failed to get users count", "error", err, "id", requestID)
		c.String(http.StatusInternalServerError, "failed to get users count")
		return
	}

	urlsCount, err := h.service.GetTotalURLsCount(c.Request.Context())
	if err != nil {
		logger.Errorw("failed to get URLs count", "error", err, "id", requestID)
		c.String(http.StatusInternalServerError, "failed to get URLs count")
		return
	}

	resp := Response{
		URLsCount:  urlsCount,
		UsersCount: usersCount,
	}
	logger.Infow("responding to request", "id", requestID, "response", resp)
	c.JSON(http.StatusOK, resp)
}

func (h *handler) isTrustedSubnet(c *gin.Context) bool {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	logger.Infow("checking subnet", "id", requestID)

	// Получаем IP-адрес из заголовка X-Real-IP
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP == "" {
		logger.Warnw("X-Real-IP header is missing", "id", requestID)
		return false
	}

	// Если доверенная подсеть не задана, разрешаем доступ
	if h.trustedSubnet == "" {
		logger.Warnw("trusted subnet is not configured", "id", requestID)
		return true
	}

	// Парсим IP-адрес клиента
	clientIP := net.ParseIP(xRealIP)
	if clientIP == nil {
		logger.Warnw("invalid IP address in X-Real-IP header", "ip", xRealIP, "id", requestID)
		return false
	}

	// Парсим CIDR подсеть
	_, ipNet, err := net.ParseCIDR(h.trustedSubnet)
	if err != nil {
		logger.Errorw("invalid trusted subnet CIDR", "subnet", h.trustedSubnet, "error", err, "id", requestID)
		return false
	}

	// Проверяем, входит ли IP-адрес в доверенную подсеть
	isInSubnet := ipNet.Contains(clientIP)

	logger.Infow("subnet check result",
		"ip", xRealIP,
		"subnet", h.trustedSubnet,
		"allowed", isInSubnet,
		"id", requestID)

	return isInSubnet
}
