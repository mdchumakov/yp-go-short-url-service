package json

import (
	"fmt"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

type creatingShortURLsAPIHandler struct {
	service service.URLShortenerService
	baseURL string
}

// NewCreatingShortURLsAPIHandler создает новый обработчик для создания короткой ссылки через JSON API.
// Принимает сервис сокращения URL и настройки приложения, возвращает обработчик, реализующий интерфейс Handler.
func NewCreatingShortURLsAPIHandler(
	service service.URLShortenerService,
	settings *config.Settings,
) handler.Handler {
	return &creatingShortURLsAPIHandler{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

// Handle CreateShortURL godoc
// @Summary Создать короткую ссылку
// @Description Создает короткую ссылку из длинного URL
// @Tags shortener
// @Accept json
// @Produce json
// @Param request body CreatingShortURLsDTOIn true "Данные для создания короткой ссылки"
// @Success 201 {object} CreatingShortURLsDTOOut "Короткая ссылка успешно создана"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 409 {object} CreatingShortURLsDTOOut "URL уже существует в системе"
// @Failure 415 {object} map[string]interface{} "Неподдерживаемый тип контента"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/shorten [post]
func (h *creatingShortURLsAPIHandler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	logger.Infow("Received shorten URL request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remote_addr", c.Request.RemoteAddr,
		"request_id", requestID)

	var dtoIn CreatingShortURLsDTOIn

	contentType := c.GetHeader("Content-Type")

	if contentType == "" {
		logger.Warnw("Missing Content-Type header",
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type header is required",
		})
		return
	} else if !strings.HasPrefix(contentType, "application/json") {
		logger.Warnw("Invalid Content-Type header",
			"content_type", contentType,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type: `application/json` header is required",
		})
		return
	}

	err := c.ShouldBindJSON(&dtoIn)
	if err != nil {
		logger.Warnw("Invalid JSON in request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	longURL := strings.TrimSpace(dtoIn.URL)
	if len(longURL) == 0 {
		logger.Warnw("Empty URL in request",
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	logger.Infow("Processing URL shortening request",
		"long_url", longURL,
		"request_id", requestID)

	shortedURL, err := h.service.ShortURL(c.Request.Context(), longURL)
	if err != nil {
		if service.IsAlreadyExistsError(err) && shortedURL != "" {
			logger.Warnw("URL already exists in storage",
				"long_url", longURL,
			)
			resultURL := h.buildShortURL(shortedURL)
			dtoOut := CreatingShortURLsDTOOut{Result: resultURL}
			c.JSON(http.StatusConflict, dtoOut)
			return
		}
		logger.Errorw("Failed to shorten URL",
			"error", err,
			"long_url", longURL,
			"request_id", requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resultURL := h.buildShortURL(shortedURL)
	logger.Infow("URL shortened successfully",
		"long_url", longURL,
		"short_url", shortedURL,
		"result_url", resultURL,
		"request_id", requestID)

	dtoOut := CreatingShortURLsDTOOut{Result: resultURL}
	c.JSON(http.StatusCreated, dtoOut)
}

func (h *creatingShortURLsAPIHandler) buildShortURL(shortedURL string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.baseURL, "/"),
		shortedURL,
	)
}
