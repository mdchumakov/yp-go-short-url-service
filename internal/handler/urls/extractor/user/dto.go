package user

// UserURLResponse представляет ответ с URL пользователя
// @Description Ответ с URL пользователя
type UserURLResponse struct {
	// @Description Сокращенный URL пользователя
	// @Example http://localhost:8080/abc123
	ShortURL string `json:"short_url" example:"http://localhost:8080/abc123"`

	// @Description Оригинальный длинный URL
	// @Example https://www.example.com/very/long/url/that/needs/to/be/shortened
	OriginalURL string `json:"original_url" example:"https://www.example.com/very/long/url/that/needs/to/be/shortened"`
}

// UserURLsResponse представляет массив ответов с URL пользователей
// @Description Массив URL пользователя
type UserURLsResponse []UserURLResponse
