package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthCheck struct{}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func (h *HealthCheck) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
