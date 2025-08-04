package url_extractor

import (
	"net/http"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ExtractingLongURLHandler struct {
	service service.LinkExtractorService
}

func NewExtractingFullLinkHandler(service service.LinkExtractorService) handler.Handler {
	return &ExtractingLongURLHandler{
		service: service,
	}
}

// Handle RedirectToLongURL godoc
// @Summary Перенаправление на длинный URL
// @Description Перенаправляет пользователя на оригинальный длинный URL по короткой ссылке
// @Tags redirect
// @Accept plain
// @Produce plain
// @Param shortURL path string true "Короткий URL" example(abc123)
// @Success 307 {string} string "Перенаправление на длинный URL"
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /{shortURL} [get]
func (h *ExtractingLongURLHandler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c)

	shortURL := c.Param("shortURL")
	if shortURL == "" {
		logger.Errorw(
			"Параметр shortURL не может быть пустым",
			"short_url", shortURL,
			"request_id", requestID)
		c.String(http.StatusBadRequest, "Параметр shortURL не может быть пустым")
		return
	}

	longURL, err := h.service.ExtractLongURL(c.Request.Context(), shortURL)
	if err != nil {
		logger.Errorw(
			"Ошибка при извлечении длинной ссылки",
			"error", err,
			"short_url", shortURL,
			"request_id", requestID,
		)
		c.String(http.StatusInternalServerError, "Ошибка при извлечении длинной ссылки: %v", err)
		return
	}
	if len(longURL) == 0 {
		logger.Infow("Ссылка не найдена в базе данных",
			"short_url", shortURL,
			"request_id", requestID,
		)
		c.String(http.StatusBadRequest, "Ссылка не найдена")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
	logger.Infow("Перенаправление на длинный URL", "redirect_url", longURL, "request_id", requestID)
}
