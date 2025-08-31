package user

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

func NewExtractingUserURLsHandler(service service.URLExtractorService, settings *config.Settings) handler.Handler {
	return &extractingUserURLsHandler{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

type extractingUserURLsHandler struct {
	baseURL string
	service service.URLExtractorService
}

// Handle GetUserURLs godoc
// @Summary Получить URL пользователя
// @Description Возвращает все URL пользователя. Требует JWT аутентификации. JWT токен должен быть передан через заголовок Authorization в формате 'Bearer <token>' или через куки с именем 'token'.
// @Tags user
// @Accept json
// @Produce json
// @Param Authorization header string false "JWT токен в заголовке Authorization (Bearer <token>)"
// @Success 200 {array} user.UserURLResponse "Список URL пользователя успешно получен"
// @Success 204 {array} user.UserURLResponse "У пользователя нет URL"
// @Failure 401 {object} object "Не авторизован - JWT токен отсутствует или недействителен"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/user/urls [get]
func (h *extractingUserURLsHandler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())
	user := middleware.GetJWTUserFromContext(c.Request.Context())
	if user == nil {
		logger.Errorw("User not found in context",
			"request_id", requestID,
		)
		c.JSON(http.StatusUnauthorized, gin.H{})

		return
	}

	userURLs, err := h.service.ExtractUserURLs(c.Request.Context(), user.ID)
	if err != nil {
		logger.Errorw("Failed to extract user URLs",
			"request_id", requestID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if len(userURLs) == 0 {
		c.JSON(http.StatusNoContent, make(UserURLsResponse, 0))
		return
	}

	// Преобразуем данные в нужный формат
	response := make(UserURLsResponse, len(userURLs))
	for i, url := range userURLs {
		response[i] = UserURLResponse{
			ShortURL:    h.buildShortURL(url.ShortURL),
			OriginalURL: url.LongURL,
		}
	}

	c.JSON(http.StatusOK, response)
	logger.Infow("Successfully returned user URLs",
		"request_id", requestID,
		"user_id", user.ID,
		"urls_count", len(userURLs),
	)
}

func (h *extractingUserURLsHandler) buildShortURL(shortedURL string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.baseURL, "/"),
		shortedURL,
	)
}
