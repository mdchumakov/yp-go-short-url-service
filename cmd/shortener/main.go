package main

import (
	"yp-go-short-url-service/internal/app"
)

func main() {
	service := app.NewApp()

	service.SetupRoutes()
	err := service.Run()
	if err != nil {
		panic(err)
	}
}
