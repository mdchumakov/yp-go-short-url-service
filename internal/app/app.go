package app

import (
	"context"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	ShortenAPI "yp-go-short-url-service/internal/handler/api/shorten"
	"yp-go-short-url-service/internal/handler/health"
	URLExtractorHandler "yp-go-short-url-service/internal/handler/urlExtractor"
	URLShortenerHandler "yp-go-short-url-service/internal/handler/urlShortener"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/middleware/gzip"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/postgres"
	HealthService "yp-go-short-url-service/internal/service/health"
	URLExtractorService "yp-go-short-url-service/internal/service/urlExtractor"
	URLShortenerService "yp-go-short-url-service/internal/service/urlShortener"

	_ "yp-go-short-url-service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
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

	// Пытаемся подключиться к PostgreSQL
	pgPool, err := db.InitPostgresDB(
		ctx,
		settings.GetDatabaseDSN(),
	)

	var repoURLs repository.URLRepository
	// PostgreSQL доступен, запускаем миграции
	err = db.RunMigrations(logger, pgPool, "migrations")
	if err != nil {
		logger.Warnw("PostgreSQL недоступен, переключаемся на SQLite", "error", err)

		// Инициализируем SQLite
		sqliteDB, err := db.SetupSQLiteDB("db/sqlite.db", logger)
		if err != nil {
			logger.Fatalw("не удалось инициализировать SQLite", "error", err)
		}

		repoURLs = db.NewSQLiteRepository(sqliteDB)
		logger.Info("SQLite успешно инициализирован как fallback")
	} else {
		repoURLs = postgres.NewURLsRepository(pgPool)
		logger.Info("PostgreSQL успешно инициализирован")
	}

	pingService := HealthService.NewHealthCheckService(repoURLs)
	urlShortenerService := URLShortenerService.NewLinkShortenerService(repoURLs)
	urlExtractorService := URLExtractorService.NewLinkExtractorService(repoURLs)

	urlExtractorHandler := URLExtractorHandler.NewExtractingFullLinkHandler(urlExtractorService)
	urlShortenerHandler := URLShortenerHandler.NewCreatingShortLinksHandler(urlShortenerService, settings)
	urlShortenerAPIHandler := ShortenAPI.NewCreatingShortLinksAPIHandler(urlShortenerService, settings)
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
