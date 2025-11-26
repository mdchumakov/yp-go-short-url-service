package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/observer/base"

	"go.uber.org/zap"
)

// LogObserver реализует интерфейс Observer для логирования событий аудита.
// Может отправлять события в файл и/или на удаленный сервер через HTTP.
type LogObserver struct {
	id         string
	logger     *zap.SugaredLogger
	settings   *config.Settings
	httpClient *http.Client
}

// NewLogObserver создает новый наблюдатель для логирования событий аудита.
// Принимает логгер и настройки приложения, возвращает реализацию интерфейса Observer.
func NewLogObserver(logger *zap.SugaredLogger, settings *config.Settings) base.Observer[Event] {

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &LogObserver{
		id:         "audit_observer",
		logger:     logger,
		settings:   settings,
		httpClient: httpClient,
	}
}

// GetID возвращает уникальный идентификатор наблюдателя.
func (o *LogObserver) GetID() string {
	return o.id
}

// Notify обрабатывает событие аудита, отправляя его в файл и/или на удаленный сервер.
// Возвращает ошибку, если отправка не удалась.
func (o *LogObserver) Notify(_ context.Context, event Event) error {

	if auditFilePath := o.settings.GetAuditFilePath(); auditFilePath != "" {
		if err := o.notifyToFile(auditFilePath, event); err != nil {
			return err
		}
	}

	if remoteAuditURL := o.settings.GetAuditURL(); remoteAuditURL != "" {
		if err := o.notifyToRemote(remoteAuditURL, event); err != nil {
			return err
		}
	}

	o.logger.Infof("LogObserver '%s' notified", o.id)
	return nil
}

func (o *LogObserver) notifyToFile(auditFilePath string, event Event) error {
	o.logger.Infof("Audit log to file: UserID=%s, Action=%s", event.UserID, event.Action)

	file, err := os.OpenFile(auditFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		o.logger.Errorf("Failed to open audit log file: %v", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			o.logger.Errorf("Failed to close audit log file: %v", err)
		}
	}(file)

	jsonData, err := json.Marshal(event)
	if err != nil {
		o.logger.Errorf("Failed to marshal audit event: %v", err)
		return err
	}

	if _, err := file.Write(append(jsonData, '\n')); err != nil {
		o.logger.Errorf("Failed to write audit event to file: %v", err)
		return err
	}

	return nil
}

func (o *LogObserver) notifyToRemote(remoteURL string, event Event) error {
	o.logger.Infof("Audit log to remote: UserID=%s, Action=%s", event.UserID, event.Action)

	jsonData, err := json.Marshal(event)
	if err != nil {
		o.logger.Errorf("Failed to marshal audit event: %v", err)
		return err
	}
	req, err := http.NewRequest("POST", remoteURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.logger.Errorf("Failed to send audit event to remote: %v", err)
		return err
	}
	defer resp.Body.Close()

	o.logger.Infow("Response from remote audit log",
		"status_code", resp.StatusCode,
	)
	return nil
}
