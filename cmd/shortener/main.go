package main

import (
	"os"
	"os/signal"
	"syscall"
	"yp-go-short-url-service/internal/app"
	"yp-go-short-url-service/internal/config"
)

// @title           URL Shortener Service API
// @version         1.0
// @description     Сервис для сокращения длинных URL-адресов
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
func main() {
	logger, err := config.NewLogger(false)
	if err != nil {
		logger.Fatal(err)
	}
	defer config.SyncLogger(logger)

	service := app.NewApp(logger)

	service.SetupCommonMiddlewares()
	service.SetupRoutes()

	// Создаем канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		if err := service.Run(); err != nil {
			logger.Fatal(err)
		}
	}()

	// Ждем сигнала завершения
	<-sigChan
	logger.Info("Received shutdown signal")

	// Корректно останавливаем приложение
	service.Stop()
}
