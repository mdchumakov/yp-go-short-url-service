-- Откат миграции для SQLite - удаляем поле user_id
-- Создаем таблицу без поля user_id
CREATE TABLE urls_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_url TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Копируем данные из текущей таблицы (без user_id)
INSERT INTO urls_old (id, short_url, long_url, created_at, updated_at)
SELECT id, short_url, long_url, created_at, updated_at FROM urls;

-- Удаляем текущую таблицу
DROP TABLE urls;

-- Переименовываем старую таблицу
ALTER TABLE urls_old RENAME TO urls;

-- Создаем индексы
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
CREATE INDEX IF NOT EXISTS idx_urls_long_url ON urls(long_url);

