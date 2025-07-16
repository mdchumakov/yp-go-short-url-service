package db

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"os"
)

const DefaultFileStoragePath = "data/urls.json"

type URL struct {
	ID        int    `json:"id"`
	ShortURL  string `json:"short_url"`
	LongURL   string `json:"long_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type FileStorageSettings struct {
	Path string `envconfig:"FILE_STORAGE_PATH"`
}

func ExtractURLSDataFromFileStorage(filePath string, log *zap.SugaredLogger) ([]URL, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file", zap.Error(err))
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error("Error closing file", zap.Error(err))
		}
	}(file)

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Error("Error reading file", zap.Error(err))
		return nil, err
	}
	var urls []URL
	if err := json.Unmarshal(bytes, &urls); err != nil {
		log.Error("Error unmarshalling JSON", zap.Error(err))
		return nil, err
	}

	log.Info("Successfully extracted URLs from file storage", zap.Int("count", len(urls)))
	return urls, nil
}
