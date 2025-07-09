package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yp-go-short-url-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_Handle(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()
	logger, _ := config.NewLogger(false)
	defer config.SyncLogger(logger)

	healthHandler := NewHealthCheck(logger)

	r.GET("/ping", healthHandler.Handle)

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	assert.NoError(t, err)

	// Action
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"pong"}`, w.Body.String())
}
