package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	Server *ServerSettings
	SQLite *SQLiteSettings
}

type ServerSettings struct {
	ServerHost   string `envconfig:"SERVER_HOST" default:"localhost" required:"true"`
	ServerPort   int    `envconfig:"SERVER_PORT" default:"8080" required:"true"`
	ServerDomain string `envconfig:"SERVER_DOMAIN" default:"localhost" required:"true"`
}

type SQLiteSettings struct {
	SQLiteDBPath string `envconfig:"SQLITE_DB_PATH" default:"db/test.db" required:"true"`
}

func NewSettings() *Settings {
	var settings Settings

	if err := envconfig.Process("", &settings); err != nil {
		panic("Failed to load settings: " + err.Error())
	}

	return &settings
}
