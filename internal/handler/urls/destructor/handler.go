package destructor

import (
	"errors"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

// NewUsersURLsDestructorAPIHandler создает новый обработчик для удаления URL пользователя через API.
// Принимает сервис удаления URL и возвращает обработчик, реализующий интерфейс Handler.
func NewUsersURLsDestructorAPIHandler(service service.URLDestructorService) handler.Handler {
	return &usersURLsDestructorAPIHandler{
		service: service,
	}
}

type usersURLsDestructorAPIHandler struct {
	service service.URLDestructorService
}

// Handle DeleteUserURLs godoc
// @Summary Удалить URL пользователя
// @Description Удаляет указанные короткие URL пользователя. Требует JWT аутентификации.
// @Tags user
// @Accept json
// @Produce json
// @Param request body DestructorRequestBodyDTOIn true "Массив коротких URL для удаления"
// @Success 202 {object} map[string]interface{} "URL успешно удалены"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 415 {object} map[string]interface{} "Неподдерживаемый тип контента"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/user/urls [delete]
func (h *usersURLsDestructorAPIHandler) Handle(c *gin.Context) {
	requestCtx := c.Request.Context()

	logger := middleware.GetLogger(requestCtx)
	requestID := middleware.ExtractRequestID(requestCtx)

	logger.Infow("Received delete URLs request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remote_addr", c.Request.RemoteAddr,
		"request_id", requestID)

	// Проверяем аутентификацию пользователя
	user := middleware.GetJWTUserFromContext(requestCtx)
	if user == nil {
		logger.Errorw("User not found in context",
			"request_id", requestID,
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// Валидируем запрос
	if err := h.validateRequest(c); err != nil {
		return
	}

	// Парсим и валидируем DTO
	dtoIn, err := h.parseAndValidateDTO(c)
	if err != nil {
		return
	}

	// Удаляем URL'ы
	err = h.service.DeleteURLsByBatch(requestCtx, dtoIn)
	if err != nil {
		logger.Errorw("Failed to delete URLs",
			"error", err,
			"user_id", user.ID,
			"request_id", requestID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete URLs",
		})
		return
	}

	logger.Infow("URLs deleted successfully",
		"user_id", user.ID,
		"urls_count", len(dtoIn),
		"request_id", requestID)

	c.JSON(http.StatusAccepted, gin.H{
		"message":       "URLs deleted successfully",
		"deleted_count": len(dtoIn),
	})
}

// validateRequest выполняет валидацию входящего запроса
func (h *usersURLsDestructorAPIHandler) validateRequest(c *gin.Context) error {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	// Проверяем Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		logger.Warnw("Missing Content-Type header",
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Content-Type header is required",
		})
		return errors.New("Content-Type header is required")
	}

	if !strings.HasPrefix(contentType, "application/json") {
		logger.Warnw("Invalid Content-Type header",
			"content_type", contentType,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Content-Type: application/json header is required",
		})
		return errors.New("Content-Type header is invalid")
	}

	return nil
}

// parseAndValidateDTO парсит и валидирует DTO из тела запроса
func (h *usersURLsDestructorAPIHandler) parseAndValidateDTO(c *gin.Context) (DestructorRequestBodyDTOIn, error) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	var dtoIn DestructorRequestBodyDTOIn

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&dtoIn); err != nil {
		logger.Warnw("Invalid JSON in request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON in request body",
		})
		return nil, err
	}

	// Валидируем DTO
	if err := dtoIn.Validate(); err != nil {
		logger.Warnw("Invalid request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return nil, err
	}

	return dtoIn, nil
}
