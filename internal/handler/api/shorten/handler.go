package shorten

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/service"
)

type CreatingShortLinksAPI struct {
	service service.LinkShortener
	baseURL string
}

func NewCreatingShortLinksAPI(service service.LinkShortener, settings *config.Settings) *CreatingShortLinksAPI {
	return &CreatingShortLinksAPI{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

func (h *CreatingShortLinksAPI) Handle(c *gin.Context) {
	var dtoIn CreatingShortLinksDTOIn

	contentType := c.GetHeader("Content-Type")

	if contentType == "" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type header is required",
		})
		return
	} else if !strings.HasPrefix(contentType, "application/json") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type: `application/json` header is required",
		})
		return
	}

	err := c.ShouldBindJSON(&dtoIn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	longURL := strings.TrimSpace(dtoIn.URL)
	if len(longURL) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	shortedURL, err := h.service.ShortenURL(longURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resultURL := h.buildShortURL(shortedURL)

	dtoOut := CreatingShortLinksDTOOut{Result: resultURL}

	c.JSON(http.StatusCreated, dtoOut)
}

func (h *CreatingShortLinksAPI) buildShortURL(shortedURL string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.baseURL, "/"),
		shortedURL,
	)
}
