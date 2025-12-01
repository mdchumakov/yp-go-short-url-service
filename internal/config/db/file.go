package db

// DefaultFileStoragePath определяет путь по умолчанию для файлового хранилища URL.
const DefaultFileStoragePath = "data/urls.json"

// FileStorageSettings содержит настройки файлового хранилища.
// Определяет путь к файлу для сохранения данных в формате JSON.
type FileStorageSettings struct {
	Path string `envconfig:"FILE_STORAGE_PATH"`
}
