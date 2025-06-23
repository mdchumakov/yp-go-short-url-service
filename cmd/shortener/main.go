package main

import (
	"flag"
	"strings"
	"yp-go-short-url-service/internal/app"
)

func init() {

}

func main() {
	service := app.NewApp()
	settings := service.GetSettings()

	service.SetupRoutes()

	connectionAddr := flag.String(
		"a",
		"",
		"Отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8888)",
	)
	redirectURL := flag.String(
		"b",
		"",
		"Отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8888)",
	)
	flag.Parse()

	if strings.TrimSpace(*redirectURL) != "" {
		settings.Server.RedirectURL = *redirectURL
	}
	err := service.Run(connectionAddr)

	if err != nil {
		panic(err)
	}
}
