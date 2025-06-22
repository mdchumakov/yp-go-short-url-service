package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/service"
)

type App struct {
	router            *gin.Engine
	shortLinksHandler *handler.CreatingShortLinks
	fullLinkHandler   *handler.ExtractingFullLink
	pingHandler       *handler.HealthCheck
	settings          *config.Settings
}

func NewApp() *App {
	router := gin.Default()
	settings := config.NewSettings()
	sqliteDB, err := db.InitSQLiteDB(settings.SQLite.SQLiteDBPath)
	if err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}

	linkShortenerService := service.NewLinkShortenerService(sqliteDB)
	handlerForCreatingShortLinks := handler.NewCreatingShortLinks(linkShortenerService, settings.Server)
	handlerForExtractingFullLink := handler.NewExtractingFullLink(sqliteDB)
	handlerHealth := handler.NewHealthCheck()

	return &App{
		router:            router,
		shortLinksHandler: handlerForCreatingShortLinks,
		fullLinkHandler:   handlerForExtractingFullLink,
		pingHandler:       handlerHealth,
		settings:          settings,
	}
}

func (a *App) SetupRoutes() {
	a.router.GET("/ping", a.pingHandler.Handle)
	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.POST("/", a.shortLinksHandler.Handle)
}

func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.settings.Server.ServerHost, a.settings.Server.ServerPort)
	fmt.Println(a.settings.Server.ServerHost)
	return a.router.Run(addr)
}
