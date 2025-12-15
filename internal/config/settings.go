package config

import (
	"fmt"
	"strings"
	"yp-go-short-url-service/internal/config/db"

	"github.com/kelseyhightower/envconfig"
)

// Settings представляет основные настройки приложения.
// Содержит настройки окружения и флаги командной строки.
type Settings struct {
	EnvSettings *ENVSettings
	Flags       *Flags
}

// ENVSettings содержит настройки, загружаемые из переменных окружения.
// Включает настройки сервера, базы данных, файлового хранилища, аудита и JWT.
type ENVSettings struct {
	Server      *ServerSettings
	PG          *db.PGSettings
	SQLite      *db.SQLiteSettings
	FileStorage *db.FileStorageSettings
	Audit       *AuditSettings
	JWT         *JWTSettings
}

// NewSettings создает новый экземпляр настроек приложения.
// Загружает настройки из переменных окружения и флагов командной строки.
func NewSettings() *Settings {
	envSettings := NewENVSettings()
	flags := NewFlags()

	return &Settings{
		EnvSettings: envSettings,
		Flags:       flags,
	}
}

// NewENVSettings создает новый экземпляр настроек окружения.
// Загружает настройки из переменных окружения с использованием envconfig.
// Паникует, если не удается загрузить настройки.
func NewENVSettings() *ENVSettings {
	var settings ENVSettings

	if err := envconfig.Process("", &settings); err != nil {
		panic("Failed to load settings: " + err.Error())
	}

	return &settings
}

// GetServerAddress возвращает адрес сервера для запуска HTTP-сервера.
// Приоритет: переменная окружения > флаг командной строки > значения по умолчанию.
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

// GetBaseURL возвращает базовый URL для генерации коротких ссылок.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
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

// GetFileStoragePath возвращает путь к файлу для хранения данных в формате JSON.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func (s *Settings) GetFileStoragePath() string {
	// Если указана переменная окружения, то используется она
	if fileStoragePath := strings.TrimSpace(s.EnvSettings.FileStorage.Path); fileStoragePath != "" {
		return fileStoragePath
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if fileStoragePath := strings.TrimSpace(s.Flags.FileStoragePath); fileStoragePath != "" {
		return fileStoragePath
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return db.DefaultFileStoragePath
}

// GetPostgresDSN возвращает строку подключения к PostgreSQL базе данных.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func (s *Settings) GetPostgresDSN() string {
	// Если указана переменная окружения, то используется она
	if dsn := strings.TrimSpace(s.EnvSettings.PG.DSN); dsn != "" {
		return dsn
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if dsn := strings.TrimSpace(s.Flags.DatabaseDSN); dsn != "" {
		return dsn
	}

	// Если нет ни переменной окружения, ни флага, то возвращается DefaultPostgresDSN
	return db.DefaultPostgresDSN
}

// GetAuditFilePath возвращает путь к файлу для сохранения логов аудита.
// Приоритет: переменная окружения > флаг командной строки > пустая строка (аудит отключен).
func (s *Settings) GetAuditFilePath() string {
	// Если указана переменная окружения, то используется она
	if auditFilePath := strings.TrimSpace(s.EnvSettings.Audit.File); auditFilePath != "" {
		return auditFilePath
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if auditFilePath := strings.TrimSpace(s.Flags.AuditFile); auditFilePath != "" {
		return auditFilePath
	}

	// Если параметр не передан, аудит в файл должен быть отключён.
	return defaultAuditFilePath
}

// GetAuditURL возвращает URL удаленного сервера для отправки логов аудита.
// Приоритет: переменная окружения > флаг командной строки > пустая строка (аудит отключен).
func (s *Settings) GetAuditURL() string {
	// Если указана переменная окружения, то используется она
	if auditURL := strings.TrimSpace(s.EnvSettings.Audit.URL); auditURL != "" {
		return auditURL
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if auditURL := strings.TrimSpace(s.Flags.AuditURL); auditURL != "" {
		return auditURL
	}

	// Если параметр не передан, аудит на удалённый сервер должен быть отключён.
	return defaultAuditURL
}

// IsHTTPSEnabled возвращает true, если HTTPS включен в настройках.
func (s *Settings) IsHTTPSEnabled() bool {
	// Если указана переменная окружения, то используется она
	if enableHTTPS := s.EnvSettings.Server.EnableHTTPS; enableHTTPS {
		return enableHTTPS
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if enableHTTPS := s.Flags.EnableHTTPS; enableHTTPS {
		return enableHTTPS
	}

	// Если параметр не передан, то не используем параметр по умолчанию.
	return defaultHTTPSUsage
}
