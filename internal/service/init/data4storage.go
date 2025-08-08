package init

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
	"yp-go-short-url-service/internal/utils"

	"go.uber.org/zap"
)

func NewDataInitializerService(repo repository.URLRepository, logger *zap.SugaredLogger) service.DataInitializerService {
	return &dataInitializerService{repo: repo, logger: logger}
}

type dataInitializerService struct {
	repo   repository.URLRepository
	logger *zap.SugaredLogger
}

func (d *dataInitializerService) Setup(ctx context.Context, fileStoragePath string) error {
	if isFileExists := utils.CheckFileExists(fileStoragePath); !isFileExists {
		urls, err := d.extractURLSDataFromRepo(ctx)
		if err != nil {
			d.logger.Errorw("Failed to extract URLs from DB", "error", err)
			return err
		}

		if err := d.saveURLSDataToFileStorage(fileStoragePath, urls); err != nil {
			d.logger.Errorw("Failed to save URLs to file storage", "error", err)
			return err
		}
		d.logger.Infow("Successfully saved URLs to file storage!")
	} else {
		urlData, err := d.extractURLSDataFromFileStorage(fileStoragePath)
		if err != nil {
			return err
		}
		if err := d.saveURLsDataToDB(ctx, urlData); err != nil {
			d.logger.Error("Error loading data into DB", zap.Error(err))
			return err
		}
	}

	return nil
}

func (d *dataInitializerService) extractURLSDataFromRepo(ctx context.Context) ([]*model.URLsModel, error) {
	var urls []*model.URLsModel

	limit, offset := 100, 0

	for {
		batch, err := d.repo.GetAll(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		urls = append(urls, batch...)
		offset += limit
	}

	return urls, nil
}

func (d *dataInitializerService) saveURLSDataToFileStorage(
	filePath string,
	urls []*model.URLsModel,
) error {
	file, err := os.Create(filePath)
	if err != nil {
		d.logger.Errorw("Error creating", "file", zap.Error(err))
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			d.logger.Errorw("Error closing", "file", zap.Error(err))
		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Устанавливаем отступы: пустая строка для префикса, 2 пробела для отступа
	if err := encoder.Encode(urls); err != nil {
		d.logger.Error("Error encoding URL to JSON", zap.Error(err))
		return err
	}
	d.logger.Debug("Successfully encoded URL to JSON")

	d.logger.Info("Successfully saved URLs to file storage")
	return nil
}

func (d *dataInitializerService) extractURLSDataFromFileStorage(filePath string) ([]*model.URLsModel, error) {
	d.logger.Info("Extracting URLs from file storage at ", zap.String("filePath", filePath))

	file, err := os.Open(filePath)
	if err != nil {
		d.logger.Error("Error opening file", zap.Error(err))
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			d.logger.Error("Error closing file", zap.Error(err))
		}
	}(file)

	bytes, err := io.ReadAll(file)
	if err != nil {
		d.logger.Error("Error reading file", zap.Error(err))
		return nil, err
	}
	var urls []*model.URLsModel
	if err := json.Unmarshal(bytes, &urls); err != nil {
		d.logger.Error("Error unmarshalling JSON", zap.Error(err))
		return nil, err
	}

	d.logger.Info("Successfully extracted URLs from file storage")
	return urls, nil
}

func (d *dataInitializerService) saveURLsDataToDB(ctx context.Context, urls []*model.URLsModel) error {
	successCount, skippedCount := 0, 0

	for _, url := range urls {
		if err := d.repo.Create(ctx, url); err != nil {
			switch {
			case repository.IsExistsError(err):
				// URL уже существует, пропускаем
				skippedCount++
				d.logger.Debugw("URL already exists, skipping",
					"short_url", url.ShortURL,
					"long_url", url.LongURL)
				continue
			default:
				d.logger.Errorw("Failed to save URL to DB",
					"error", err,
					"short_url", url.ShortURL,
					"long_url", url.LongURL)
				return err
			}
		}

		successCount++
		d.logger.Debugw("Successfully saved URL to DB",
			"short_url", url.ShortURL,
			"long_url", url.LongURL)
	}

	d.logger.Infow("Finished saving URLs to DB",
		"total", len(urls),
		"success", successCount,
		"skipped", skippedCount,
	)

	return nil
}
