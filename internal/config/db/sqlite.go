package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type SQLiteSettings struct {
	SQLiteDBPath string `envconfig:"SQLITE_DB_PATH" default:"db/test.db" required:"true"`
}

// InitSQLiteDB инициализирует соединение с SQLite базой данных
func InitSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return db, nil
}

// SetupSQLiteDB настраивает SQLite базу данных и создает таблицы
func SetupSQLiteDB(dbPath string, log *zap.SugaredLogger) (*sql.DB, error) {
	db, err := InitSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	// Создаем таблицу urls
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_url TEXT NOT NULL UNIQUE,
		long_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create urls table: %w", err)
	}

	// Создаем таблицу users
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		password TEXT,
		is_anonymous BOOLEAN DEFAULT FALSE,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createUsersTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	// Создаем таблицу user_urls
	createUserUrlsTableSQL := `
	CREATE TABLE IF NOT EXISTS user_urls (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		url_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(createUserUrlsTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create user_urls table: %w", err)
	}

	// Создаем индексы для улучшения производительности
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);",
		"CREATE INDEX IF NOT EXISTS idx_urls_long_url ON urls(long_url);",
		"CREATE INDEX IF NOT EXISTS idx_user_urls_user_id ON user_urls(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_urls_url_id ON user_urls(url_id);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_urls_user_id_url_id ON user_urls(user_id, url_id);",
	}

	for _, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			log.Warnf("Failed to create index: %v", err)
		}
	}

	log.Info("Successfully initialized SQLite database")
	return db, nil
}
