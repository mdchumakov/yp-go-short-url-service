-- Создаем таблицу пользователей для SQLite
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    password TEXT,
    is_anonymous INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индексы для таблицы пользователей
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_users_is_anonymous ON users(is_anonymous);

-- Добавляем поле user_id в таблицу urls для SQLite
-- SQLite не поддерживает ALTER TABLE ADD COLUMN напрямую, поэтому создаем новую таблицу
CREATE TABLE urls_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_url TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    user_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Копируем данные из старой таблицы
INSERT INTO urls_new (id, short_url, long_url, created_at, updated_at)
SELECT id, short_url, long_url, created_at, updated_at FROM urls;

-- Удаляем старую таблицу
DROP TABLE urls;

-- Переименовываем новую таблицу
ALTER TABLE urls_new RENAME TO urls;

-- Создаем индексы
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
CREATE INDEX IF NOT EXISTS idx_urls_long_url ON urls(long_url);
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);

