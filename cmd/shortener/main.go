package main

import (
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

	err = service.Run()

	if err != nil {
		logger.Fatal(err)
	}
}
