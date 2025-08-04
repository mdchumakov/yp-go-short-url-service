package db

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"os"
)

const DefaultFileStoragePath = "data/urls.json"

type URL struct {
	ID        int    `json:"ID"`
	ShortURL  string `json:"ShortURL"`
	LongURL   string `json:"LongURL"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

type FileStorageSettings struct {
	Path string `envconfig:"FILE_STORAGE_PATH"`
}

func ExtractURLSDataFromFileStorage(filePath string, log *zap.SugaredLogger) ([]URL, error) {
	log.Info("Extracting URLs from file storage at ", zap.String("filePath", filePath))

	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file", zap.Error(err))
		return nil, err
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

func SaveURLSDataToFileStorage(filePath string, urls []interface{}, log *zap.SugaredLogger) error {
	log.Info("Saving URLs to file storage at ", zap.String("filePath", filePath))
	file, err := os.Create(filePath)
	if err != nil {
		log.Error("Error creating file", zap.Error(err))
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error("Error closing file", zap.Error(err))
		}
	}(file)

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(urls); err != nil {
		log.Error("Error encoding URL to JSON", zap.Error(err))
		return err
	}
	log.Debug("Successfully encoded URL to JSON")

	log.Info("Successfully saved URLs to file storage", zap.Int("count", len(urls)))
	return nil
}
