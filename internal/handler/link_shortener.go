package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/service"
)

const maxBodySize int64 = 1024 * 1024 // 1 MB

type CreatingShortLinks struct {
	service service.LinkShortener
	baseURL string
}

func NewCreatingShortLinksHandler(service service.LinkShortener, settings *config.Settings) *CreatingShortLinks {
	return &CreatingShortLinks{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

func (h *CreatingShortLinks) Handle(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	if contentType == "" {
		c.String(http.StatusBadRequest, "Отсутствует Content-Type")
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
	if err != nil {
		c.String(http.StatusBadRequest, "Ошибка чтения данных")
		return
	}

	longURL := strings.TrimSpace(string(body))
	if len(longURL) == 0 {
		c.String(http.StatusBadRequest, "Пустой URL")
		return
	}

	shortedURL, err := h.service.ShortenURL(longURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка при сокращении URL: %v", err)
		return
	}

	resultURL := h.buildShortURL(shortedURL)
	c.String(http.StatusCreated, resultURL)
}

func (h *CreatingShortLinks) buildShortURL(shortedURL string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.baseURL, "/"),
		shortedURL,
	)
}
