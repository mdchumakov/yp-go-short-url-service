package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/handler/api/shorten"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"
)

type App struct {
	router               *gin.Engine
	shortLinksHandler    *handler.CreatingShortLinks
	shortLinksHandlerAPI *shorten.CreatingShortLinksAPI
	fullLinkHandler      *handler.ExtractingFullLink
	pingHandler          *handler.HealthCheck
	settings             *config.Settings
	logger               *zap.SugaredLogger
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
	handlerForCreatingShortLinksAPI := shorten.NewCreatingShortLinksAPI(linkShortenerService, settings)
	handlerForExtractingFullLink := handler.NewExtractingFullLink(sqliteDB)
	handlerHealth := handler.NewHealthCheck(logger)

	return &App{
		router:               router,
		shortLinksHandler:    handlerForCreatingShortLinks,
		shortLinksHandlerAPI: handlerForCreatingShortLinksAPI,
		fullLinkHandler:      handlerForExtractingFullLink,
		pingHandler:          handlerHealth,
		settings:             settings,
		logger:               logger,
	}
}

func (a *App) SetupMiddlewares() {
	a.router.Use(middleware.LoggerMiddleware(a.logger))
	a.router.Use(middleware.GZIPMiddleware(a.logger))
}

func (a *App) SetupRoutes() {
	a.router.GET("/ping", a.pingHandler.Handle)
	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.POST("/", a.shortLinksHandler.Handle)
	a.router.POST("/api/shorten", a.shortLinksHandlerAPI.Handle)
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}
