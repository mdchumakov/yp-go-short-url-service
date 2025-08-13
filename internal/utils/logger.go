package utils

import (
	"context"
	"yp-go-short-url-service/internal/middleware"

	"go.uber.org/zap"
)

// LoggerWrapper - обертка для логгера с дополнительными методами
type LoggerWrapper struct {
	logger *zap.SugaredLogger
}

// NewLoggerWrapper создает новую обертку для логгера
func NewLoggerWrapper(logger *zap.SugaredLogger) *LoggerWrapper {
	return &LoggerWrapper{logger: logger}
}

// FromContext создает обертку из логгера в контексте
func NewLoggerWrapperFromContext(ctx context.Context) *LoggerWrapper {
	return &LoggerWrapper{logger: middleware.GetLogger(ctx)}
}

// WithRequestID добавляет request_id к логгеру
func (l *LoggerWrapper) WithRequestID(ctx context.Context) *zap.SugaredLogger {
	requestID := middleware.ExtractRequestID(ctx)
	return l.logger.With("request_id", requestID)
}

// Info логирует информационное сообщение
func (l *LoggerWrapper) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.WithRequestID(ctx).Infow(msg, keysAndValues...)
}

// Error логирует ошибку
func (l *LoggerWrapper) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.WithRequestID(ctx).Errorw(msg, keysAndValues...)
}

// Debug логирует отладочное сообщение
func (l *LoggerWrapper) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.WithRequestID(ctx).Debugw(msg, keysAndValues...)
}
