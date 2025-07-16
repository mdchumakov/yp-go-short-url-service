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
	log.Info("Extracting URLs from file storage at ", zap.String("filePath", filePath))
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Error("Error getting file info", zap.Error(err))
		return nil, err
	}
	if fileInfo.IsDir() {
		log.Error("File storage path is a directory", zap.String("path", filePath))
		files, err := os.ReadDir(filePath)
		if err != nil {
			log.Error("Error reading directory", zap.Error(err))
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() {
				log.Warn("Skipping directory in file storage", zap.String("directory", file.Name()))
				continue
			}
			filePath = filePath + "/" + file.Name()
			log.Info("Using file from directory", zap.String("filePath", filePath))
			break
		}
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("File not found, using default file storage path", zap.String("defaultPath", DefaultFileStoragePath))
			file, err = os.Open(DefaultFileStoragePath)
			if err != nil {
				log.Error("Error opening default file storage", zap.Error(err))
				return nil, err
			}
		} else {
			return nil, err
		}
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
