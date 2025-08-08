package db

const DefaultFileStoragePath = "data/urls.json"

type FileStorageSettings struct {
	Path string `envconfig:"FILE_STORAGE_PATH"`
}
