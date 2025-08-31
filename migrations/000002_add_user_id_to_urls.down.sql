-- Удаляем индексы
DROP INDEX IF EXISTS idx_user_urls_user_id_url_id;
DROP INDEX IF EXISTS idx_user_urls_url_id;
DROP INDEX IF EXISTS idx_user_urls_user_id;

-- Удаляем таблицы
DROP TABLE IF EXISTS user_urls CASCADE;
DROP TABLE IF EXISTS users CASCADE;
