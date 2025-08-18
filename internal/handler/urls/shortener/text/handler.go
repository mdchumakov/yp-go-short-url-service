package text

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"
	"yp-go-short-url-service/internal/service/urls/shortener"

	"github.com/gin-gonic/gin"
)

const maxBodySize int64 = 1024 * 1024 // 1 MB

type CreatingShortLinks struct {
	service service.LinkShortenerService
	baseURL string
}

func NewCreatingShortLinksHandler(
	service service.LinkShortenerService,
	settings *config.Settings,
) handler.Handler {
	return &CreatingShortLinks{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

// Handle CreateShortURLText godoc
// @Summary Создать короткую ссылку (текстовый формат)
// @Description Создает короткую ссылку из длинного URL, отправленного в текстовом формате
// @Tags shortener
// @Accept plain
// @Produce plain
// @Param url body string true "Длинный URL для сокращения"
// @Success 201 {string} string "Сокращенный URL"
// @Failure 400 {string} string "Неверный запрос"
// @Failure 409 {string} string "URL уже существует в системе"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router / [post]
func (h *CreatingShortLinks) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())
	user := middleware.GetJWTUserFromContext(c.Request.Context())
	if user == nil {
		logger.Errorw("JWT user is nil", "request_id", requestID)
		c.String(http.StatusInternalServerError, "Ошибка аутентификации пользователя")
		return
	}

	contentType := c.GetHeader("Content-Type")

	if contentType == "" {
		logger.Infow("Content-Type header is missing", "request_id", requestID)
		c.String(http.StatusBadRequest, "Отсутствует Content-Type")
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
	if err != nil {
		logger.Errorw("Failed to read request body", "body", string(body), "error", err, "request_id", requestID)
		c.String(http.StatusBadRequest, "Ошибка чтения данных")
		return
	}

	longURL := strings.TrimSpace(string(body))
	if len(longURL) == 0 {
		logger.Warnw("Received empty URL", "request_id", requestID)
		c.String(http.StatusBadRequest, "Пустой URL")
		return
	}

	shortedURL, err := h.service.ShortURL(c.Request.Context(), longURL)
	if err != nil {
		if shortener.IsAlreadyExistsError(err) && shortedURL != "" {
			logger.Warnw("URL already exists in storage", "long_url", longURL, "request_id", requestID)
			resultURL := h.buildShortURL(shortedURL)
			c.String(http.StatusConflict, resultURL)
			return
		}
		logger.Errorw("Failed to shorten URL", "long_url", longURL, "error", err, "request_id", requestID)
		c.String(http.StatusInternalServerError, "Ошибка при сокращении URL: %v", err)
		return
	}

	resultURL := h.buildShortURL(shortedURL)
	c.String(http.StatusCreated, resultURL)
	logger.Infow("URL shortened successfully", "long_url", longURL, "short_url", shortedURL, "result_url", resultURL, "request_id", requestID)
}

func (h *CreatingShortLinks) buildShortURL(shortedURL string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.baseURL, "/"),
		shortedURL,
	)
}
