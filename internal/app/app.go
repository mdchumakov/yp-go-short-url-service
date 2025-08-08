package app

import (
	"context"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/handler/health"
	urlExtractorHandler "yp-go-short-url-service/internal/handler/urls/extractor"
	shortenBatchAPI "yp-go-short-url-service/internal/handler/urls/shortener/batch"
	shortenAPI "yp-go-short-url-service/internal/handler/urls/shortener/json"
	urlShortenerHandler "yp-go-short-url-service/internal/handler/urls/shortener/text"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/middleware/gzip"
	baseRepo "yp-go-short-url-service/internal/repository/base"
	healthService "yp-go-short-url-service/internal/service/health"
	initService "yp-go-short-url-service/internal/service/init"
	urlExtractorService "yp-go-short-url-service/internal/service/urls/extractor"
	urlShortenerService "yp-go-short-url-service/internal/service/urls/shortener"

	_ "yp-go-short-url-service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type App struct {
	router                    *gin.Engine
	shortLinksHandler         handler.Handler
	shortLinksHandlerAPI      handler.Handler
	shortLinksBatchHandlerAPI handler.Handler
	fullLinkHandler           handler.Handler
	pingHandler               handler.Handler
	settings                  *config.Settings
	logger                    *zap.SugaredLogger
}

func NewApp(logger *zap.SugaredLogger) *App {
	ctx := context.Background()

	router := gin.Default()
	settings := config.NewSettings()

	dbPool := db.Setup(ctx, logger, &db.SetupParams{
		PostgresDSN:      settings.GetPostgresDSN(),
		SQLiteDSN:        settings.EnvSettings.SQLite.SQLiteDBPath,
		PGMigrationsPath: settings.EnvSettings.PG.MigrationsPath,
	})
	repoURLs := baseRepo.NewURLsRepository(dbPool)
	InitService := initService.NewDataInitializerService(repoURLs, logger)
	if err := InitService.Setup(ctx, settings.GetFileStoragePath()); err != nil {
		logger.Fatalw("Failed to initialize data", "error", err)
	}

	pingService := healthService.NewHealthCheckService(repoURLs)
	URLShortenerService := urlShortenerService.NewURLShortenerService(repoURLs)
	URLExtractorService := urlExtractorService.NewLinkExtractorService(repoURLs)

	URLExtractorHandler := urlExtractorHandler.NewExtractingFullLinkHandler(URLExtractorService)
	URLShortenerHandler := urlShortenerHandler.NewCreatingShortLinksHandler(URLShortenerService, settings)
	URLShortenerAPIHandler := shortenAPI.NewCreatingShortURLsAPIHandler(URLShortenerService, settings)
	URLShortenerBatchAPIHandler := shortenBatchAPI.NewCreatingShortURLsByBatchAPIHandler(URLShortenerService, settings)
	healthHandler := health.NewPingHandler(pingService)

	return &App{
		router:                    router,
		shortLinksHandler:         URLShortenerHandler,
		shortLinksHandlerAPI:      URLShortenerAPIHandler,
		shortLinksBatchHandlerAPI: URLShortenerBatchAPIHandler,
		fullLinkHandler:           URLExtractorHandler,
		pingHandler:               healthHandler,
		settings:                  settings,
		logger:                    logger,
	}
}

func (a *App) SetupMiddlewares() {
	a.router.Use(middleware.RequestIDMiddleware(a.logger))
	a.router.Use(middleware.LoggerMiddleware(a.logger))
	a.router.Use(gzip.Middleware(a.logger))
}

func (a *App) SetupRoutes() {
	a.router.GET("/ping", a.pingHandler.Handle)
	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.POST("/", a.shortLinksHandler.Handle)
	a.router.POST("/api/shorten", a.shortLinksHandlerAPI.Handle)
	a.router.POST("/api/shorten/batch", a.shortLinksBatchHandlerAPI.Handle)
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}
