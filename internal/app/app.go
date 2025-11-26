package app

import (
	"context"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	"yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/handler/health"
	urlsDestructorAPIHandler "yp-go-short-url-service/internal/handler/urls/destructor"
	urlExtractorHandler "yp-go-short-url-service/internal/handler/urls/extractor"
	userURLsHandler "yp-go-short-url-service/internal/handler/urls/extractor/user"
	shortenBatchAPI "yp-go-short-url-service/internal/handler/urls/shortener/batch"
	shortenAPI "yp-go-short-url-service/internal/handler/urls/shortener/json"
	urlShortenerHandler "yp-go-short-url-service/internal/handler/urls/shortener/text"
	"yp-go-short-url-service/internal/observer"
	"yp-go-short-url-service/internal/observer/audit"

	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/middleware/gzip"
	baseRepo "yp-go-short-url-service/internal/repository/base"
	"yp-go-short-url-service/internal/service"
	authService "yp-go-short-url-service/internal/service/auth"
	healthService "yp-go-short-url-service/internal/service/health"
	initService "yp-go-short-url-service/internal/service/init"
	jwtService "yp-go-short-url-service/internal/service/jwt"
	urlDestructorService "yp-go-short-url-service/internal/service/urls/destructor"
	urlExtractorService "yp-go-short-url-service/internal/service/urls/extractor"
	urlShortenerService "yp-go-short-url-service/internal/service/urls/shortener"

	_ "yp-go-short-url-service/docs"

	"net/http"
	"net/http/pprof"

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
	destructorAPIHandler      handler.Handler
	fullLinkHandler           handler.Handler
	userURLsHandler           handler.Handler
	pingHandler               handler.Handler
	services                  Services
	settings                  *config.Settings
	logger                    *zap.SugaredLogger
}

type Services struct {
	auth          service.AuthService
	jwt           service.JWTService
	urlDestructor service.URLDestructorService
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

	auditEventBus := observer.NewEventBus[audit.Event]()

	auditLogObserver := audit.NewLogObserver(logger, settings)
	auditEventBus.Subscribe(auditLogObserver)

	pingService := healthService.NewHealthCheckService(repoURLs)
	URLShortenerService := urlShortenerService.NewURLShortenerService(repoURLs, userURLsRepo, auditEventBus)
	URLExtractorService := urlExtractorService.NewLinkExtractorService(repoURLs, userURLsRepo, auditEventBus)
	URLDestructorService := urlDestructorService.NewURLDestructorService(repoURLs, userURLsRepo)

	URLExtractorHandler := urlExtractorHandler.NewExtractingFullLinkHandler(URLExtractorService)
	UserURLsHandler := userURLsHandler.NewExtractingUserURLsHandler(URLExtractorService, settings)
	URLShortenerHandler := urlShortenerHandler.NewCreatingShortLinksHandler(URLShortenerService, settings)
	URLShortenerAPIHandler := shortenAPI.NewCreatingShortURLsAPIHandler(URLShortenerService, settings)
	URLShortenerBatchAPIHandler := shortenBatchAPI.NewCreatingShortURLsByBatchAPIHandler(URLShortenerService, settings)
	URLDestructorAPIHandler := urlsDestructorAPIHandler.NewUsersURLsDestructorAPIHandler(URLDestructorService)
	HealthHandler := health.NewPingHandler(pingService)

	return &App{
		router:                    router,
		shortLinksHandler:         URLShortenerHandler,
		shortLinksHandlerAPI:      URLShortenerAPIHandler,
		shortLinksBatchHandlerAPI: URLShortenerBatchAPIHandler,
		destructorAPIHandler:      URLDestructorAPIHandler,
		fullLinkHandler:           URLExtractorHandler,
		userURLsHandler:           UserURLsHandler,
		pingHandler:               HealthHandler,
		services: Services{
			auth:          AuthService,
			jwt:           JWTService,
			urlDestructor: URLDestructorService,
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
	anonNotAllowedMiddleware := middleware.JWTAuthMiddleware(a.services.jwt, a.services.auth, a.settings.EnvSettings.JWT, false, a.logger)

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
		privateGroup.DELETE("/api/user/urls", a.destructorAPIHandler.Handle)
	}

	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Добавляем pprof роуты для профилирования
	a.setupPprofRoutes()
}

func (a *App) Run() error {
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}

// setupPprofRoutes настраивает роуты для pprof профилирования
func (a *App) setupPprofRoutes() {
	pprofGroup := a.router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapH(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
		pprofGroup.GET("/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
		pprofGroup.GET("/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
		pprofGroup.GET("/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
		pprofGroup.GET("/heap", gin.WrapH(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/goroutine", gin.WrapH(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/allocs", gin.WrapH(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/block", gin.WrapH(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/mutex", gin.WrapH(http.HandlerFunc(pprof.Index)))
	}
}

// Stop - корректно останавливает приложение
func (a *App) Stop() {
	a.logger.Info("Stopping application...")
	if a.services.urlDestructor != nil {
		a.services.urlDestructor.Stop()
	}
	a.logger.Info("Application stopped")
}
