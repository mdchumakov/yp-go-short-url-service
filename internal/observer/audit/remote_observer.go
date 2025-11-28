package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"yp-go-short-url-service/internal/observer/base"

	"go.uber.org/zap"
)

// RemoteObserver реализует интерфейс Observer для отправки событий аудита на удаленный сервер.
type RemoteObserver struct {
	id         string
	logger     *zap.SugaredLogger
	remoteURL  string
	httpClient *http.Client
}

// NewRemoteObserver создает новый наблюдатель для отправки событий аудита на удаленный сервер.
// Принимает логгер и URL удаленного сервера, возвращает реализацию интерфейса Observer.
func NewRemoteObserver(logger *zap.SugaredLogger, remoteURL string) base.Observer[Event] {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &RemoteObserver{
		id:         "audit_remote_observer",
		logger:     logger,
		remoteURL:  remoteURL,
		httpClient: httpClient,
	}
}

// GetID возвращает уникальный идентификатор наблюдателя.
func (o *RemoteObserver) GetID() string {
	return o.id
}

// Notify обрабатывает событие аудита, отправляя его на удаленный сервер через HTTP.
// Возвращает ошибку, если отправка не удалась.
func (o *RemoteObserver) Notify(_ context.Context, event Event) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		o.logger.Errorf("Failed to marshal audit event: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", o.remoteURL, strings.NewReader(string(jsonData)))
	if err != nil {
		o.logger.Errorf("Failed to create HTTP request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.logger.Errorf("Failed to send audit event to remote: %v", err)
		return err
	}
	// CI ругается если замкну вызов
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		o.logger.Warnw("Unexpected status code from remote audit log",
			"status_code", resp.StatusCode,
		)
	}

	o.logger.Debugf("RemoteObserver '%s' notified: UserID=%s, Action=%s, StatusCode=%d",
		o.id, event.UserID, event.Action, resp.StatusCode)
	return nil
}

// Stop выполняет остановку наблюдателя.
func (o *RemoteObserver) Stop() error {
	o.logger.Infof("RemoteObserver '%s' stopped", o.id)
	return nil
}
