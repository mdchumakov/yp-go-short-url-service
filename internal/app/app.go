package app

import (
	"context"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	shorten_api "yp-go-short-url-service/internal/handler/api/shorten"
	"yp-go-short-url-service/internal/handler/health"
	"yp-go-short-url-service/internal/handler/url_extractor"
	"yp-go-short-url-service/internal/handler/url_shortener"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/middleware/gzip"
	"yp-go-short-url-service/internal/repository"
	health_ "yp-go-short-url-service/internal/service/health"
	urlExtractor "yp-go-short-url-service/internal/service/url_extractor"
	urlShortener "yp-go-short-url-service/internal/service/url_shortener"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	_ "yp-go-short-url-service/docs"
)

type App struct {
	router               *gin.Engine
	shortLinksHandler    handler.Handler
	shortLinksHandlerAPI handler.Handler
	fullLinkHandler      handler.Handler
	pingHandler          handler.Handler
	settings             *config.Settings
	logger               *zap.SugaredLogger
}

func NewApp(logger *zap.SugaredLogger) *App {
	ctx := context.Background()

	router := gin.Default()
	settings := config.NewSettings()

	pgPool, err := db.InitPostgresDB(
		ctx,
		settings.GetDatabaseDSN(),
	)
	if err != nil {
		logger.Fatal(err)
	}

	if err = db.RunMigrations(logger, pgPool, "./migrations"); err != nil {
		logger.Fatal("failed to run migrations", "error", err)
	}

	repoURLs := repository.NewURLsRepository(pgPool)

	pingService := health_.NewHealthCheckService(repoURLs)
	urlShortenerService := urlShortener.NewLinkShortenerService(repoURLs)
	urlExtractorService := urlExtractor.NewLinkExtractorService(repoURLs)

	urlExtractorHandler := url_extractor.NewExtractingFullLinkHandler(urlExtractorService)
	urlShortenerHandler := url_shortener.NewCreatingShortLinksHandler(urlShortenerService, settings)
	urlShortenerAPIHandler := shorten_api.NewCreatingShortLinksAPIHandler(urlShortenerService, settings)
	healthHandler := health.NewPingHandler(pingService)

	return &App{
		router:               router,
		shortLinksHandler:    urlShortenerHandler,
		shortLinksHandlerAPI: urlShortenerAPIHandler,
		fullLinkHandler:      urlExtractorHandler,
		pingHandler:          healthHandler,
		settings:             settings,
		logger:               logger,
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
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}
