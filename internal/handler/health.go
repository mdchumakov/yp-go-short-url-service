package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type HealthCheck struct {
	logger *zap.SugaredLogger
}

func NewHealthCheck(logger *zap.SugaredLogger) *HealthCheck {
	return &HealthCheck{logger: logger}
}

func (h *HealthCheck) Handle(c *gin.Context) {
	h.logger.Infow("starting health check")
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
	h.logger.Infow("health check done")
}
