package health

import (
	"net/http"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

type pingHandler struct {
	service service.HealthCheckService
}

// NewPingHandler создает новый обработчик для проверки здоровья сервиса.
// Принимает сервис проверки здоровья и возвращает обработчик, реализующий интерфейс Handler.
func NewPingHandler(service service.HealthCheckService) handler.Handler {
	return &pingHandler{
		service: service,
	}
}

// Handle Ping godoc
// @Summary Проверка здоровья сервиса
// @Description Проверяет доступность базы данных и возвращает статус сервиса
// @Tags health
// @Accept plain
// @Produce plain
// @Success 200 {string} string "pong"
// @Failure 500 {string} string "health check failed"
// @Router /ping [get]
func (h *pingHandler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	logger.Infow("Starting health check", "requestID", requestID)
	if err := h.service.Ping(c.Request.Context()); err != nil {
		logger.Errorw("Health check failed",
			"error", err,
			"request_id", requestID,
		)
		c.String(http.StatusInternalServerError, `{"error": "health check failed"}`)
		return
	}

	c.String(http.StatusOK, `pong`)
	logger.Infow("Health check done", "requestID", requestID)
}
