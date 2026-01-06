package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"yp-go-short-url-service/internal/app"
	"yp-go-short-url-service/internal/config"
)

var buildVersion, buildDate, buildCommit string

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
	printMetaInfo()

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	logger, err := config.NewLogger(false)
	if err != nil {
		fmt.Printf("failed to initialize logger: %s\n", err)
		os.Exit(1)
	}
	defer config.SyncLogger(logger)

	service, err := app.NewApp(logger, ctx)
	if err != nil {
		logger.Errorw("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	service.SetupCommonMiddlewares()
	service.SetupRoutes()

	// Запускаем сервер в горутине
	go func() {
		if err := service.Run(); err != nil {
			logger.Fatal(err)
		}
	}()

	// Ждем сигнала завершения через контекст
	<-ctx.Done()
	logger.Info("Received shutdown signal")

	// Корректно останавливаем приложение
	service.Stop()
}

func printMetaInfo() {
	var defaultInfo = "N/A"

	if buildVersion == "" {
		buildVersion = defaultInfo
	}

	if buildDate == "" {
		buildDate = defaultInfo
	}

	if buildCommit == "" {
		buildCommit = defaultInfo
	}

	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}
