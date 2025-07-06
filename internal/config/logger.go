package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

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
			logger.Fatal("Failed to sync logger", "error", err)
		}
	}
}
