package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"
)

type App struct {
	router            *gin.Engine
	shortLinksHandler *handler.CreatingShortLinks
	fullLinkHandler   *handler.ExtractingFullLink
	pingHandler       *handler.HealthCheck
	settings          *config.Settings
	logger            *zap.SugaredLogger
}

func NewApp(logger *zap.SugaredLogger) *App {
	router := gin.Default()
	settings := config.NewSettings()

	sqliteDB, err := db.InitSQLiteDB(settings.EnvSettings.SQLite.SQLiteDBPath)
	if err != nil {
		logger.Fatal(err)
	}

	linkShortenerService := service.NewLinkShortenerService(sqliteDB)
	handlerForCreatingShortLinks := handler.NewCreatingShortLinksHandler(linkShortenerService, settings)
	handlerForExtractingFullLink := handler.NewExtractingFullLink(sqliteDB)
	handlerHealth := handler.NewHealthCheck(logger)

	return &App{
		router:            router,
		shortLinksHandler: handlerForCreatingShortLinks,
		fullLinkHandler:   handlerForExtractingFullLink,
		pingHandler:       handlerHealth,
		settings:          settings,
		logger:            logger,
	}
}

func (a *App) SetupMiddlewares() {
	a.router.Use(middleware.LoggerMiddleware(a.logger))
}

func (a *App) SetupRoutes() {
	a.router.GET("/ping", a.pingHandler.Handle)
	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.POST("/", a.shortLinksHandler.Handle)
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}
