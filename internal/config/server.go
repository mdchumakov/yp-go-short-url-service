package config

const (
	defaultServerHost = "localhost"
	defaultServerPort = 8080
	defaultBaseURL    = "http://localhost:8080/"
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
}
