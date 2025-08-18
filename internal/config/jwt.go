package config

import (
	"net/http"
	"time"
)

// JWTSettings содержит настройки для JWT токенов
type JWTSettings struct {
	SecretKey     string        `envconfig:"JWT_SECRET_KEY" default:"" required:"false"`
	TokenDuration time.Duration `envconfig:"JWT_TOKEN_DURATION" default:"24h" required:"false"`
	Issuer        string        `envconfig:"JWT_ISSUER" default:"yp-go-short-url-service" required:"false"`
	Algorithm     string        `envconfig:"JWT_ALGORITHM" default:"HS256" required:"false"`

	// Настройки куки
	CookieName     string `envconfig:"JWT_COOKIE_NAME" default:"token" required:"false"`
	CookiePath     string `envconfig:"JWT_COOKIE_PATH" default:"/" required:"false"`
	CookieDomain   string `envconfig:"JWT_COOKIE_DOMAIN" default:"" required:"false"`
	CookieSecure   bool   `envconfig:"JWT_COOKIE_SECURE" default:"false" required:"false"`
	CookieHttpOnly bool   `envconfig:"JWT_COOKIE_HTTP_ONLY" default:"true" required:"false"`
	CookieSameSite string `envconfig:"JWT_COOKIE_SAME_SITE" default:"lax" required:"false"`
}

// GetCookieSameSite возвращает http.SameSite значение на основе строковой настройки
func (j *JWTSettings) GetCookieSameSite() http.SameSite {
	switch j.CookieSameSite {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
