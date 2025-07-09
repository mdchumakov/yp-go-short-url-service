package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"strings"
	"yp-go-short-url-service/internal/config/db"
)

type Settings struct {
	EnvSettings *ENVSettings
	Flags       *Flags
}

type ENVSettings struct {
	Server *ServerSettings
	SQLite *db.SQLiteSettings
}

func NewSettings() *Settings {
	envSettings := NewENVSettings()
	flags := NewFlags()

	return &Settings{
		EnvSettings: envSettings,
		Flags:       flags,
	}
}

func NewENVSettings() *ENVSettings {
	var settings ENVSettings

	if err := envconfig.Process("", &settings); err != nil {
		panic("Failed to load settings: " + err.Error())
	}

	return &settings
}

func (s *Settings) GetServerAddress() string {
	// Если указана переменная окружения, то используется она
	if serverAddr := strings.TrimSpace(s.EnvSettings.Server.ServerAddress); serverAddr != "" {
		return serverAddr
	}
	if strings.TrimSpace(s.EnvSettings.Server.ServerHost) != "" &&
		s.EnvSettings.Server.ServerPort != 0 {
		return fmt.Sprintf(
			"%s:%d",
			s.EnvSettings.Server.ServerHost,
			s.EnvSettings.Server.ServerPort,
		)
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if serverAddr := strings.TrimSpace(s.Flags.ServerAddress); serverAddr != "" {
		return serverAddr
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return fmt.Sprintf(
		"%s:%d",
		defaultServerHost,
		defaultServerPort,
	)
}

func (s *Settings) GetBaseURL() string {
	// Если указана переменная окружения, то используется она
	if baseURL := strings.TrimSpace(s.EnvSettings.Server.BaseURL); baseURL != "" {
		return baseURL
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if baseURL := strings.TrimSpace(s.Flags.BaseURL); baseURL != "" {
		return baseURL
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return defaultBaseURL
}
