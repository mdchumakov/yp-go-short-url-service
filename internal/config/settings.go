package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/samber/lo"

	"strings"
	"yp-go-short-url-service/internal/config/db"
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
	GRPCAddress     string `json:"grpc_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	AuditFilePath   string `json:"audit_file_path"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
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
		b, err := os.ReadFile(jsonConfigPath)
		if err != nil {
			return nil, fmt.Errorf("could not read json config file %s: %w", jsonConfigPath, err)
		}

		var parsedJSON SettingsFromJSON
		if err = json.Unmarshal(b, &parsedJSON); err != nil {
			return nil, fmt.Errorf("could not parse json config file %s: %w", jsonConfigPath, err)
		}
		fileConfig = &parsedJSON
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
	var envServerAddr, flagServerAddr, confServerAddr string
	var envServerPort, flagServerPort, confServerPort int

	if s.EnvSettings != nil {
		envServerAddr = strings.TrimSpace(s.EnvSettings.Server.ServerAddress)
		envServerPort = s.EnvSettings.Server.ServerPort
	}

	if s.Flags != nil {
		flagServerAddr = strings.TrimSpace(s.Flags.ServerAddress)
	}

	if s.JSONConfig != nil {
		confServerAddr = strings.TrimSpace(s.JSONConfig.ServerAddress)
	}

	return fmt.Sprintf(
		"%s:%d",
		lo.CoalesceOrEmpty(envServerAddr, flagServerAddr, confServerAddr, defaultServerHost),
		lo.CoalesceOrEmpty(envServerPort, flagServerPort, confServerPort, defaultServerPort),
	)
}

func (s *Settings) GetGRPCServerAddress() string {
	var envGRPCServerAddr, flagGRPCServerAddr, confGRPCServerAddr string
	var envGRPCServerPort, flagGRPCServerPort, confGRPCServerPort int

	if s.EnvSettings != nil {
		if serverAddr := strings.TrimSpace(s.EnvSettings.Server.GRPCAddress); serverAddr != "" {
			return serverAddr
		}
		envGRPCServerHost := strings.TrimSpace(s.EnvSettings.Server.GRPCHost)
		envGRPCServerPort = s.EnvSettings.Server.ServerPort

		if envGRPCServerHost != "" && envGRPCServerPort != 0 {
			envGRPCServerAddr = fmt.Sprintf("%s:%d", envGRPCServerHost, envGRPCServerPort)
		}
	}

	if s.Flags != nil {
		flagGRPCServerAddr = strings.TrimSpace(s.Flags.GRPCAddress)
	}

	if s.JSONConfig != nil {
		confGRPCServerAddr = strings.TrimSpace(s.JSONConfig.GRPCAddress)
	}

	return fmt.Sprintf(
		"%s:%d",
		lo.CoalesceOrEmpty(envGRPCServerAddr, flagGRPCServerAddr, confGRPCServerAddr, defaultGRPCHost),
		lo.CoalesceOrEmpty(envGRPCServerPort, flagGRPCServerPort, confGRPCServerPort, defaultGRPCPort),
	)
}

// GetBaseURL возвращает базовый URL для генерации коротких ссылок.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func (s *Settings) GetBaseURL() string {
	var envBaseURL, flagBaseURL, confBaseURL string

	if s.EnvSettings != nil {
		envBaseURL = strings.TrimSpace(s.EnvSettings.Server.BaseURL)
	}

	if s.Flags != nil {
		flagBaseURL = strings.TrimSpace(s.Flags.BaseURL)
	}

	if s.JSONConfig != nil {
		confBaseURL = strings.TrimSpace(s.JSONConfig.BaseURL)
	}

	return lo.CoalesceOrEmpty(
		envBaseURL,
		flagBaseURL,
		confBaseURL,
		defaultBaseURL,
	)
}

// GetFileStoragePath возвращает путь к файлу для хранения данных в формате JSON.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func (s *Settings) GetFileStoragePath() string {
	var envFileStoragePath, flagFileStoragePath, confFileStoragePath string

	if s.EnvSettings != nil {
		envFileStoragePath = strings.TrimSpace(s.EnvSettings.FileStorage.Path)
	}

	if s.Flags != nil {
		flagFileStoragePath = strings.TrimSpace(s.Flags.FileStoragePath)
	}

	if s.JSONConfig != nil {
		confFileStoragePath = strings.TrimSpace(s.JSONConfig.FileStoragePath)
	}

	return lo.CoalesceOrEmpty(
		envFileStoragePath,
		flagFileStoragePath,
		confFileStoragePath,
		db.DefaultFileStoragePath,
	)
}

// GetPostgresDSN возвращает строку подключения к PostgreSQL базе данных.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func (s *Settings) GetPostgresDSN() string {
	var envDSN, flagDSN, confDSN string

	if s.EnvSettings != nil {
		envDSN = strings.TrimSpace(s.EnvSettings.PG.DSN)
	}

	if s.Flags != nil {
		flagDSN = strings.TrimSpace(s.Flags.DatabaseDSN)
	}

	if s.JSONConfig != nil {
		confDSN = strings.TrimSpace(s.JSONConfig.DatabaseDSN)
	}

	return lo.CoalesceOrEmpty(envDSN, flagDSN, confDSN, db.DefaultPostgresDSN)
}

// GetAuditFilePath возвращает путь к файлу для сохранения логов аудита.
// Приоритет: переменная окружения > флаг командной строки > пустая строка (аудит отключен).
func (s *Settings) GetAuditFilePath() string {
	var envAuditPath, flagAuditPath, confAuditPath string

	if s.EnvSettings != nil {
		envAuditPath = strings.TrimSpace(s.EnvSettings.Audit.File)
	}

	if s.Flags != nil {
		flagAuditPath = strings.TrimSpace(s.Flags.AuditFile)
	}

	if s.JSONConfig != nil {
		confAuditPath = strings.TrimSpace(s.JSONConfig.AuditFilePath)
	}
	return lo.CoalesceOrEmpty(envAuditPath, flagAuditPath, confAuditPath, defaultAuditFilePath)
}

// GetAuditURL возвращает URL удаленного сервера для отправки логов аудита.
// Приоритет: переменная окружения > флаг командной строки > пустая строка (аудит отключен).
func (s *Settings) GetAuditURL() string {
	var envAuditURL, flagAuditURL, confAuditURL string

	if s.EnvSettings != nil {
		envAuditURL = strings.TrimSpace(s.EnvSettings.Audit.URL)
	}

	if s.Flags != nil {
		flagAuditURL = strings.TrimSpace(s.Flags.AuditURL)
	}

	return lo.CoalesceOrEmpty(envAuditURL, flagAuditURL, confAuditURL, defaultAuditURL)
}

// IsHTTPSEnabled возвращает true, если HTTPS включен в настройках.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию (false).
func (s *Settings) IsHTTPSEnabled() bool {
	var envEnableHTTPS, flagEnableHTTPS, confEnableHTTPS bool

	if s.EnvSettings != nil {
		envEnableHTTPS = s.EnvSettings.Server.EnableHTTPS
	}

	if s.Flags != nil {
		flagEnableHTTPS = s.Flags.EnableHTTPS
	}

	if s.JSONConfig != nil {
		confEnableHTTPS = s.JSONConfig.EnableHTTPS
	}

	return lo.CoalesceOrEmpty(envEnableHTTPS, flagEnableHTTPS, confEnableHTTPS, defaultHTTPSUsage)
}

func (s *Settings) GetTrustedSubnet() string {
	var envTrustedSubnet, flagTrustedSubnet, confTrustedSubnet string

	if s.EnvSettings != nil {
		envTrustedSubnet = strings.TrimSpace(s.EnvSettings.Server.TrustedSubnet)
	}

	if s.Flags != nil {
		flagTrustedSubnet = strings.TrimSpace(s.Flags.TrustedSubnet)
	}

	if s.JSONConfig != nil {
		confTrustedSubnet = strings.TrimSpace(s.JSONConfig.TrustedSubnet)
	}

	return lo.CoalesceOrEmpty(
		envTrustedSubnet,
		flagTrustedSubnet,
		confTrustedSubnet,
		defaultTrustedSubnet,
	)
}
