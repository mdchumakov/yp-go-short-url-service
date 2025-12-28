package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/samber/lo"

	"strings"
	"yp-go-short-url-service/internal/config/db"
	ufiles "yp-go-short-url-service/internal/utils/files"
)

// Settings представляет основные настройки приложения.
// Содержит настройки окружения и флаги командной строки.
type Settings struct {
	EnvSettings *ENVSettings
	Flags       *Flags
	JSONConfig  *SettingsFromJSON
}

// ENVSettings содержит настройки, загружаемые из переменных окружения.
// Включает настройки сервера, базы данных, файлового хранилища, аудита и JWT.
type ENVSettings struct {
	Server         *ServerSettings
	PG             *db.PGSettings
	SQLite         *db.SQLiteSettings
	FileStorage    *db.FileStorageSettings
	Audit          *AuditSettings
	JWT            *JWTSettings
	ConfigJSONPath string `envconfig:"CONFIG" default:"" required:"false"`
}

// SettingsFromJSON используется для загрузки настроек из JSON-файла.
// Включает настройки сервера, базы данных, файлового хранилища, аудита и JWT.
type SettingsFromJSON struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	AuditFilePath   string `json:"audit_file_path"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// NewSettings создает новый экземпляр настроек приложения.
// Загружает настройки из переменных окружения и флагов командной строки.
// Если путь к JSON-файлу указан, но файл не может быть прочитан или распарсен,
// возвращает ошибку.
func NewSettings() (*Settings, error) {
	envSettings := NewENVSettings()
	flags := NewFlags()

	// Определяем путь к JSON-файлу конфигурации
	jsonConfigPath := lo.CoalesceOrEmpty(flags.JSONConfigPath, envSettings.ConfigJSONPath)

	var fileConfig *SettingsFromJSON
	// Если путь к JSON-файлу указан, пытаемся его распарсить
	// Если путь не указан, fileConfig остается nil (это нормально)
	if jsonConfigPath != "" {
		var err error
		fileConfig, err = ufiles.ParseJSON[SettingsFromJSON](jsonConfigPath)
		if err != nil {
			return nil, fmt.Errorf("could not parse json config: %w", err)
		}
	}

	return &Settings{
		EnvSettings: envSettings,
		Flags:       flags,
		JSONConfig:  fileConfig,
	}, nil
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

	return fmt.Sprintf(
		"%s:%d",
		lo.CoalesceOrEmpty(s.JSONConfig.ServerAddress, defaultServerHost),
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

	return lo.CoalesceOrEmpty(s.JSONConfig.BaseURL, defaultBaseURL)
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

	return lo.CoalesceOrEmpty(s.JSONConfig.FileStoragePath, db.DefaultFileStoragePath)
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

	return lo.CoalesceOrEmpty(s.JSONConfig.DatabaseDSN, db.DefaultPostgresDSN)
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

	return lo.CoalesceOrEmpty(s.JSONConfig.AuditFilePath, defaultAuditFilePath)
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
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию (false).
func (s *Settings) IsHTTPSEnabled() bool {
	// Если указана переменная окружения, то используется она
	if enableHTTPS := s.EnvSettings.Server.EnableHTTPS; enableHTTPS {
		return enableHTTPS
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if enableHTTPS := s.Flags.EnableHTTPS; enableHTTPS {
		return enableHTTPS
	}

	return lo.CoalesceOrEmpty(s.JSONConfig.EnableHTTPS, defaultHTTPSUsage)
}
