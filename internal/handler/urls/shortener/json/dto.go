package json

// CreatingShortLinksDTOIn представляет входные данные для создания короткой ссылки
type CreatingShortLinksDTOIn struct {
	// URL - длинный URL для сокращения
	// required: true
	// example: "https://www.example.com/very/long/url/that/needs/to/be/shortened"
	URL string `json:"url" binding:"required"`
}

// CreatingShortLinksDTOOut представляет выходные данные после создания короткой ссылки
type CreatingShortLinksDTOOut struct {
	// Result - сокращенный URL
	// example: "http://localhost:8080/abc123"
	Result string `json:"result"`
}
