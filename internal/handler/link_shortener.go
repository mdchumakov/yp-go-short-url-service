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
	service        service.LinkShortener
	serverSettings *config.ServerSettings
}

func NewCreatingShortLinks(service service.LinkShortener, settings *config.ServerSettings) *CreatingShortLinks {
	return &CreatingShortLinks{service: service, serverSettings: settings}
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

	resultURL := h.buildShortURL(c, shortedURL)
	c.String(http.StatusCreated, resultURL)
}

func (h *CreatingShortLinks) buildShortURL(c *gin.Context, shortedURL string) string {
	fmt.Println(h.serverSettings)

	if redirectURL := h.serverSettings.RedirectURL; strings.TrimSpace(redirectURL) != "" {
		return fmt.Sprintf(
			"%s/%s",
			strings.TrimRight(redirectURL, "/"),
			shortedURL,
		)
	}

	scheme := "http"
	host := h.serverSettings.ServerHost
	port := fmt.Sprintf("%d", h.serverSettings.ServerPort)
	if c.Request.TLS != nil {
		scheme = "https"
		host = h.serverSettings.ServerDomain
		port = ""

		return fmt.Sprintf(
			"%s://%s/%s",
			scheme,
			host,
			shortedURL,
		)
	}

	return fmt.Sprintf(
		"%s://%s:%s/%s",
		scheme,
		host,
		port,
		shortedURL,
	)

}
