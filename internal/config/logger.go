package config

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.SugaredLogger

// InitGlobalLogger инициализирует глобальный логгер
func InitGlobalLogger(level string) error {
	config := zap.NewProductionConfig()

	// Парсим уровень логирования
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger.Sugar()
	return nil
}

// GetGlobalLogger возвращает глобальный логгер
func GetGlobalLogger() *zap.SugaredLogger {
	if globalLogger == nil {
		// Возвращаем no-op логгер если глобальный не инициализирован
		return zap.NewNop().Sugar()
	}
	return globalLogger
}

// SetGlobalLogger устанавливает глобальный логгер (для тестов)
func SetGlobalLogger(logger *zap.SugaredLogger) {
	globalLogger = logger
}

func NewLogger(isProd bool) (*zap.SugaredLogger, error) {
	if isProd {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}
		return logger.Sugar(), nil
	}

	return NewDevLogger()
}

func NewDevLogger() (*zap.SugaredLogger, error) {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.TimeKey = "T"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger.Sugar(), nil
}

func SyncLogger(logger *zap.SugaredLogger) {
	if logger != nil {
		if err := logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "bad file descriptor") {
				logger.Error("Failed to sync logger", "error", err)
			}
		}
	}
}
