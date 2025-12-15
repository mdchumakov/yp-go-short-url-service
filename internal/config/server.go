package config

const (
	defaultServerHost = "localhost"
	defaultServerPort = 8080
	defaultBaseURL    = "http://localhost:8080/"
	defaultHttpsUsage = false
)

// ServerSettings содержит настройки HTTP-сервера.
// Включает адрес, хост, порт, домен, базовый URL и окружение приложения.
type ServerSettings struct {
	ServerAddress string `envconfig:"SERVER_ADDRESS" default:"" required:"false"`
	ServerHost    string `envconfig:"SERVER_HOST" default:"" required:"false"`
	ServerPort    int    `envconfig:"SERVER_PORT" default:"0" required:"false"`
	ServerDomain  string `envconfig:"SERVER_DOMAIN" default:"localhost" required:"true"`
	BaseURL       string `envconfig:"BASE_URL" default:"" required:"false"`
	Environment   string `envconfig:"ENVIRONMENT" default:"development" required:"false"`
	EnableHTTPS   bool   `envconfig:"ENABLE_HTTPS" default:"false"`
}

// IsProd возвращает true, если текущее окружение является производственным (production).
// Проверяет значение поля Environment на "production" или "prod".
func (s *ServerSettings) IsProd() bool {
	return s.Environment == "production" || s.Environment == "prod"
}
