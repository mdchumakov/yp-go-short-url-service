package audit

import (
	"yp-go-short-url-service/internal/observer/base"

	"go.uber.org/zap"
)

// NewEventBus создает и настраивает шину событий для аудита.
// Принимает путь к файлу аудита и URL для отправки аудита, а также логгер.
// Возвращает указатель на шину событий.
func NewEventBus(auditFilePath, auditURL string, logger *zap.SugaredLogger) base.Subject[Event] {
	auditEventBus := base.NewEventBus[Event]()

	// Подписываем FileObserver, если указан путь к файлу
	if auditFilePath != "" {
		fileObserver, err := NewFileObserver(logger, auditFilePath)
		if err != nil {
			logger.Errorf("Failed to create file audit observer: %v", err)
		}
		auditEventBus.Subscribe(fileObserver)

	}

	// Подписываем RemoteObserver, если указан URL удаленного сервера
	if auditURL != "" {
		remoteObserver := NewRemoteObserver(logger, auditURL)
		auditEventBus.Subscribe(remoteObserver)
	}
	return auditEventBus
}
