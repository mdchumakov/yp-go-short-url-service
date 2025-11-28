package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"

	"yp-go-short-url-service/internal/observer/base"

	"go.uber.org/zap"
)

// FileObserver реализует интерфейс Observer для логирования событий аудита в файл.
// Использует буферизованную запись для эффективной работы с файлом.
type FileObserver struct {
	id       string
	logger   *zap.SugaredLogger
	filePath string
	file     *os.File
	writer   *bufio.Writer
	mutex    sync.Mutex
}

// NewFileObserver создает новый наблюдатель для логирования событий аудита в файл.
// Открывает файл и инициализирует буферизованный writer.
// Возвращает ошибку, если не удалось открыть файл.
func NewFileObserver(logger *zap.SugaredLogger, filePath string) (base.Observer[Event], error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Errorf("Failed to open audit log file: %v", err)
		return nil, err
	}

	writer := bufio.NewWriter(file)

	return &FileObserver{
		id:       "audit_file_observer",
		logger:   logger,
		filePath: filePath,
		file:     file,
		writer:   writer,
	}, nil
}

// GetID возвращает уникальный идентификатор наблюдателя.
func (o *FileObserver) GetID() string {
	return o.id
}

// Notify обрабатывает событие аудита, записывая его в файл.
// Использует буферизованную запись для повышения производительности.
// Возвращает ошибку, если запись не удалась.
func (o *FileObserver) Notify(_ context.Context, event Event) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	jsonData, err := json.Marshal(event)
	if err != nil {
		o.logger.Errorf("Failed to marshal audit event: %v", err)
		return err
	}

	if _, err := o.writer.Write(jsonData); err != nil {
		o.logger.Errorf("Failed to write audit event to buffer: %v", err)
		return err
	}

	if err := o.writer.WriteByte('\n'); err != nil {
		o.logger.Errorf("Failed to write newline to buffer: %v", err)
		return err
	}

	// Периодически сбрасываем буфер на диск
	if err := o.writer.Flush(); err != nil {
		o.logger.Errorf("Failed to flush audit event to file: %v", err)
		return err
	}

	o.logger.Debugf("FileObserver '%s' notified: UserID=%s, Action=%s", o.id, event.UserID, event.Action)
	return nil
}

// Stop закрывает файл и освобождает ресурсы.
// Должен быть вызван при завершении работы приложения.
func (o *FileObserver) Stop() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.writer != nil {
		if err := o.writer.Flush(); err != nil {
			o.logger.Errorf("Failed to flush buffer before closing: %v", err)
			return err
		}
	}

	if o.file != nil {
		if err := o.file.Close(); err != nil {
			o.logger.Errorf("Failed to close audit log file: %v", err)
			return err
		}
	}

	o.logger.Infof("FileObserver '%s' closed", o.id)
	return nil
}
