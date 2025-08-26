package batch

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

func NewCreatingShortURLsByBatchAPIHandler(
	service service.URLShortenerService,
	settings *config.Settings,
) handler.Handler {
	return &creatingShortURLsByBatchAPIHandler{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

type creatingShortURLsByBatchAPIHandler struct {
	service service.URLShortenerService
	baseURL string
}

// Handle CreateShortURLsByBatch godoc
// @Summary Создать короткие ссылки пакетно
// @Description Создает короткие ссылки из массива длинных URL-ов
// @Tags shortener
// @Accept json
// @Produce json
// @Param request body CreatingShortURLsByBatchDTOIn true "Массив данных для создания коротких ссылок"
// @Success 201 {array} URLResponse "Короткие ссылки успешно созданы"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 415 {object} map[string]interface{} "Неподдерживаемый тип контента"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/shorten/batch [post]
func (h *creatingShortURLsByBatchAPIHandler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	logger.Infow("Received shorten URL request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"remote_addr", c.Request.RemoteAddr,
		"request_id", requestID)

	// Валидация запроса
	if err := h.validateRequest(c); err != nil {
		return
	}

	// Обработка запроса
	if err := h.processRequest(c); err != nil {
		return
	}
}

// validateRequest выполняет валидацию входящего запроса
func (h *creatingShortURLsByBatchAPIHandler) validateRequest(c *gin.Context) error {
	if err := h.validateContentType(c); err != nil {
		return err
	}

	var dtoIn CreatingShortURLsByBatchDTOIn
	if err := h.parseAndValidateDTO(c, &dtoIn); err != nil {
		return err
	}

	// Сохраняем DTO в контексте для использования в processRequest
	c.Set("dto_in", dtoIn)
	return nil
}

// parseAndValidateDTO парсит и валидирует DTO
func (h *creatingShortURLsByBatchAPIHandler) parseAndValidateDTO(c *gin.Context, dtoIn *CreatingShortURLsByBatchDTOIn) error {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	if err := c.ShouldBindJSON(dtoIn); err != nil {
		logger.Warnw("Invalid JSON in request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	if err := h.validateDTOIn(*dtoIn); err != nil {
		logger.Warnw("Invalid request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	return nil
}

// processRequest обрабатывает валидированный запрос
func (h *creatingShortURLsByBatchAPIHandler) processRequest(c *gin.Context) error {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	// Получаем DTO из контекста
	dtoInInterface, _ := c.Get("dto_in")
	dtoIn := dtoInInterface.(CreatingShortURLsByBatchDTOIn)

	logger.Infow("Processing URL shortening request",
		"dto_in", dtoIn,
		"request_id", requestID,
	)

	// Создаем короткие URL'ы
	shortedURLs, err := h.createShortURLs(c, dtoIn)
	if err != nil {
		return err
	}

	// Формируем ответ
	return h.buildResponse(c, shortedURLs)
}

// createShortURLs создает короткие URL'ы через сервис
func (h *creatingShortURLsByBatchAPIHandler) createShortURLs(c *gin.Context, dtoIn CreatingShortURLsByBatchDTOIn) ([]map[string]string, error) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	shortedURLs, err := h.service.ShortURLsByBatch(c.Request.Context(), dtoIn.ToMapSlice())
	if err != nil {
		logger.Errorw("Failed to shorten URL",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, err
	}

	logger.Infow("URL shortened successfully")
	return shortedURLs, nil
}

// buildResponse формирует и отправляет ответ клиенту
func (h *creatingShortURLsByBatchAPIHandler) buildResponse(c *gin.Context, shortedURLs []map[string]string) error {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	// Добавляем базовый URL к коротким URL'ам
	if err := h.buildShortURL(shortedURLs); err != nil {
		logger.Errorw("Failed to build short URL",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	// Преобразуем в DTO для ответа
	dtoOut := h.convertToDTOOut(shortedURLs)

	logger.Infow("URLs shortened successfully",
		"count", len(dtoOut),
		"request_id", requestID)

	c.JSON(http.StatusCreated, dtoOut)
	return nil
}

func (h *creatingShortURLsByBatchAPIHandler) validateContentType(c *gin.Context) error {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c)

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		logger.Warnw("Missing Content-Type header",
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type header is required",
		})
		return errors.New("Content-Type header is required")
	} else if !strings.HasPrefix(contentType, "application/json") {
		logger.Warnw("Invalid Content-Type header",
			"content_type", contentType,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Content-type: `application/json` header is required",
		})
		return errors.New("Content-Type header is invalid")
	}

	return nil
}

func (h *creatingShortURLsByBatchAPIHandler) validateDTOIn(dtoIn CreatingShortURLsByBatchDTOIn) error {
	if len(dtoIn) == 0 {
		return errors.New("urls array is empty")
	}
	return nil
}

func (h *creatingShortURLsByBatchAPIHandler) buildShortURL(shortedURLs []map[string]string) error {
	var shortedURL string
	for _, row := range shortedURLs {
		shortedURL = row["short_url"]
		row["short_url"] = fmt.Sprintf("%s/%s", strings.TrimRight(h.baseURL, "/"), shortedURL)
	}
	return nil
}

// convertToDTOOut преобразует []map[string]string в CreatingShortURLsByBatchDTOOut
func (h *creatingShortURLsByBatchAPIHandler) convertToDTOOut(shortedURLs []map[string]string) CreatingShortURLsByBatchDTOOut {
	dtoOut := make(CreatingShortURLsByBatchDTOOut, len(shortedURLs))
	for i, row := range shortedURLs {
		dtoOut[i] = URLResponse{
			CorrelationID: row["correlation_id"],
			ShortURL:      row["short_url"],
		}
	}
	return dtoOut
}
