package batch

// URLRequest представляет один элемент запроса для сокращения URL
type URLRequest struct {
	// CorrelationID - уникальный идентификатор для корреляции запроса/ответа
	// required: true
	// example: "1"
	CorrelationID string `json:"correlation_id" binding:"required"`
	// OriginalURL - длинный URL для сокращения
	// required: true
	// example: "https://www.example.com/very/long/url/that/needs/to/be/shortened"
	OriginalURL string `json:"original_url" binding:"required"`
}

// CreatingShortURLsByBatchDTOIn представляет массив запросов для пакетного сокращения URL
type CreatingShortURLsByBatchDTOIn []URLRequest

// ToMapSlice преобразует CreatingShortURLsByBatchDTOIn в []map[string]string
func (dto CreatingShortURLsByBatchDTOIn) ToMapSlice() []map[string]string {
	result := make([]map[string]string, len(dto))
	for i, req := range dto {
		result[i] = map[string]string{
			"correlation_id": req.CorrelationID,
			"original_url":   req.OriginalURL,
		}
	}
	return result
}

// URLResponse represents the response containing details of a shortened URL.
// CorrelationID is a unique identifier for the request/response correlation.
// ShortURL holds the generated shortened URL string.
type URLResponse struct {
	// CorrelationID - уникальный идентификатор для корреляции запроса/ответа
	// example: "1"
	CorrelationID string `json:"correlation_id"`
	// ShortURL - сокращенный URL
	// example: "http://localhost:8080/abc123"
	ShortURL string `json:"short_url"`
}

// CreatingShortURLsByBatchDTOOut represents a slice of URLResponse objects for batch URL shortening operations.
type CreatingShortURLsByBatchDTOOut []URLResponse
