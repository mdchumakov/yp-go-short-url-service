package app

import (
	"context"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/handler/health"
	urlExtractorHandler "yp-go-short-url-service/internal/handler/urls/extractor"
	userURLsHandler "yp-go-short-url-service/internal/handler/urls/extractor/user"
	shortenBatchAPI "yp-go-short-url-service/internal/handler/urls/shortener/batch"
	shortenAPI "yp-go-short-url-service/internal/handler/urls/shortener/json"
	urlShortenerHandler "yp-go-short-url-service/internal/handler/urls/shortener/text"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/middleware/gzip"
	baseRepo "yp-go-short-url-service/internal/repository/base"
	"yp-go-short-url-service/internal/service"
	authService "yp-go-short-url-service/internal/service/auth"
	healthService "yp-go-short-url-service/internal/service/health"
	initService "yp-go-short-url-service/internal/service/init"
	jwtService "yp-go-short-url-service/internal/service/jwt"
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
	userURLsHandler           handler.Handler
	pingHandler               handler.Handler
	services                  Services
	settings                  *config.Settings
	logger                    *zap.SugaredLogger
}

type Services struct {
	auth service.AuthService
	jwt  service.JWTService
}

func NewApp(logger *zap.SugaredLogger) *App {
	ctx := context.Background()

	router := gin.Default()
	settings := config.NewSettings()
	jwtSettings := settings.EnvSettings.JWT

	dbPool := db.Setup(ctx, logger, &db.SetupParams{
		PostgresDSN:      settings.GetPostgresDSN(),
		SQLiteDSN:        settings.EnvSettings.SQLite.SQLiteDBPath,
		PGMigrationsPath: settings.EnvSettings.PG.MigrationsPath,
	})
	repoURLs := baseRepo.NewURLsRepository(dbPool)
	userRepo := baseRepo.NewUsersRepository(dbPool)
	userURLsRepo := baseRepo.NewUserURLsRepository(dbPool)
	InitService := initService.NewDataInitializerService(repoURLs, logger)
	if err := InitService.Setup(ctx, settings.GetFileStoragePath()); err != nil {
		logger.Fatalw("Failed to initialize data", "error", err)
	}

	AuthService := authService.NewAuthService(userRepo, jwtSettings)
	JWTService := jwtService.NewJWTService(jwtSettings)

	pingService := healthService.NewHealthCheckService(repoURLs)
	URLShortenerService := urlShortenerService.NewURLShortenerService(repoURLs, userURLsRepo)
	URLExtractorService := urlExtractorService.NewLinkExtractorService(repoURLs, userURLsRepo)

	URLExtractorHandler := urlExtractorHandler.NewExtractingFullLinkHandler(URLExtractorService)
	UserURLsHandler := userURLsHandler.NewExtractingUserURLsHandler(URLExtractorService, settings)
	URLShortenerHandler := urlShortenerHandler.NewCreatingShortLinksHandler(URLShortenerService, settings)
	URLShortenerAPIHandler := shortenAPI.NewCreatingShortURLsAPIHandler(URLShortenerService, settings)
	URLShortenerBatchAPIHandler := shortenBatchAPI.NewCreatingShortURLsByBatchAPIHandler(URLShortenerService, settings)
	HealthHandler := health.NewPingHandler(pingService)

	return &App{
		router:                    router,
		shortLinksHandler:         URLShortenerHandler,
		shortLinksHandlerAPI:      URLShortenerAPIHandler,
		shortLinksBatchHandlerAPI: URLShortenerBatchAPIHandler,
		fullLinkHandler:           URLExtractorHandler,
		userURLsHandler:           UserURLsHandler,
		pingHandler:               HealthHandler,
		services: Services{
			auth: AuthService,
			jwt:  JWTService,
		},
		settings: settings,
		logger:   logger,
	}
}

func (a *App) SetupCommonMiddlewares() {
	a.router.Use(middleware.RequestIDMiddleware(a.logger))
	a.router.Use(middleware.LoggerMiddleware(a.logger))
	a.router.Use(gzip.Middleware(a.logger))
}

func (a *App) SetupRoutes() {
	anonAllowedMiddleware := middleware.JWTAuthMiddleware(a.services.jwt, a.services.auth, a.settings.EnvSettings.JWT, true, a.logger)
	anonNotAllowedMiddleware := middleware.JWTAuthMiddleware(a.services.jwt, a.services.auth, a.settings.EnvSettings.JWT, true, a.logger)

	publicGroup := a.router.Group("/")
	publicGroup.Use(anonAllowedMiddleware)
	{
		publicGroup.GET("/ping", a.pingHandler.Handle)
		publicGroup.POST("/", a.shortLinksHandler.Handle)
		publicGroup.POST("/api/shorten", a.shortLinksHandlerAPI.Handle)
		publicGroup.POST("/api/shorten/batch", a.shortLinksBatchHandlerAPI.Handle)
	}

	privateGroup := a.router.Group("/")
	privateGroup.Use(anonNotAllowedMiddleware)
	{
		privateGroup.GET("/api/user/urls", a.userURLsHandler.Handle)
	}

	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}
