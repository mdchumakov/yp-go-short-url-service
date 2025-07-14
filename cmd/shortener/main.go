package main

import (
	"yp-go-short-url-service/internal/app"
	"yp-go-short-url-service/internal/config"
)

func main() {
	logger, err := config.NewLogger(false)
	if err != nil {
		logger.Fatal(err)
	}
	defer config.SyncLogger(logger)

	service := app.NewApp(logger)

	service.SetupMiddlewares()
	service.SetupRoutes()

	err = service.Run()

	if err != nil {
		logger.Fatal(err)
	}
}
