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
	service service.LinkShortenerService,
	settings *config.Settings,
) handler.Handler {
	return &creatingShortURLsByBatchAPIHandler{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

type creatingShortURLsByBatchAPIHandler struct {
	service service.LinkShortenerService
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
	if err := h.validateContentType(c); err != nil {
		return
	}

	var dtoIn CreatingShortURLsByBatchDTOIn
	if err := c.ShouldBindJSON(&dtoIn); err != nil {
		logger.Warnw("Invalid JSON in request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validateDTOIn(dtoIn); err != nil {
		logger.Warnw("Invalid request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Infow("Processing URL shortening request",
		"dto_in", dtoIn,
		"request_id", requestID,
	)

	shortedURLs, err := h.service.ShortURLsByBatch(c.Request.Context(), dtoIn.ToMapSlice())
	if err != nil {
		logger.Errorw("Failed to shorten URL",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Infow("URL shortened successfully")

	if err = h.buildShortURL(shortedURLs); err != nil {
		logger.Errorw("Failed to build short URL",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем []map[string]string в CreatingShortURLsByBatchDTOOut
	dtoOut := h.convertToDTOOut(shortedURLs)

	logger.Infow("URLs shortened successfully",
		"count", len(dtoOut),
		"request_id", requestID)

	c.JSON(http.StatusCreated, dtoOut)
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
