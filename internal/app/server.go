package app

import (
	"context"
	"fmt"
	"net"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/config/db"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/handler"
	grpcImpl "yp-go-short-url-service/internal/handler/grpc"
	"yp-go-short-url-service/internal/handler/health"
	statsHandler "yp-go-short-url-service/internal/handler/stats"
	urlsDestructorAPIHandler "yp-go-short-url-service/internal/handler/urls/destructor"
	urlExtractorHandler "yp-go-short-url-service/internal/handler/urls/extractor"
	userURLsHandler "yp-go-short-url-service/internal/handler/urls/extractor/user"
	shortenBatchAPI "yp-go-short-url-service/internal/handler/urls/shortener/batch"
	shortenAPI "yp-go-short-url-service/internal/handler/urls/shortener/json"
	urlShortenerHandler "yp-go-short-url-service/internal/handler/urls/shortener/text"
	"yp-go-short-url-service/internal/observer/audit"
	"yp-go-short-url-service/internal/observer/base"

	"yp-go-short-url-service/internal/middleware"
	grpcMiddleware "yp-go-short-url-service/internal/middleware/grpc"
	"yp-go-short-url-service/internal/middleware/gzip"
	baseRepo "yp-go-short-url-service/internal/repository/base"
	"yp-go-short-url-service/internal/service"
	authService "yp-go-short-url-service/internal/service/auth"
	healthService "yp-go-short-url-service/internal/service/health"
	initService "yp-go-short-url-service/internal/service/init"
	jwtService "yp-go-short-url-service/internal/service/jwt"
	statsService "yp-go-short-url-service/internal/service/stats"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App представляет основное приложение сервиса сокращения URL.
// Содержит роутер, обработчики запросов, сервисы и настройки.
type App struct {
	router                    *gin.Engine
	grpcServer                *grpc.Server
	shortLinksHandler         handler.Handler
	shortLinksHandlerAPI      handler.Handler
	shortLinksBatchHandlerAPI handler.Handler
	destructorAPIHandler      handler.Handler
	fullLinkHandler           handler.Handler
	userURLsHandler           handler.Handler
	pingHandler               handler.Handler
	statsHandler              handler.Handler
	services                  Services
	settings                  *config.Settings
	logger                    *zap.SugaredLogger
	dataBus                   DataBus
}

// Services содержит коллекцию сервисов приложения.
// Используется для доступа к сервисам аутентификации, JWT и удаления URL.
type Services struct {
	auth          service.AuthService
	jwt           service.JWTService
	urlDestructor service.URLDestructorService
}

// DataBus содержит все шины событий для передачи данных между компонентами приложения.
type DataBus struct {
	auditEventBus base.Subject[audit.Event]
}

// NewApp создает новый экземпляр приложения с инициализированными зависимостями.
// Инициализирует базу данных, репозитории, сервисы, обработчики и настраивает маршруты.
// Возвращает готовый к использованию экземпляр App или ошибку, если инициализация не удалась.
func NewApp(logger *zap.SugaredLogger, ctx context.Context) (*App, error) {
	router := gin.Default()
	settings, err := config.NewSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	jwtSettings := settings.EnvSettings.JWT

	dbPool := db.Setup(ctx, logger, &db.SetupParams{
		PostgresDSN:      settings.GetPostgresDSN(),
		SQLiteDSN:        settings.EnvSettings.SQLite.SQLiteDBPath,
		PGMigrationsPath: settings.EnvSettings.PG.MigrationsPath,
	})
	// Проверяем, что db.Setup не вернул ошибку
	if err, ok := dbPool.(error); ok {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	repoURLs := baseRepo.NewURLsRepository(dbPool)
	userRepo := baseRepo.NewUsersRepository(dbPool)
	userURLsRepo := baseRepo.NewUserURLsRepository(dbPool)
	InitService := initService.NewDataInitializerService(repoURLs, logger)
	if err := InitService.Setup(ctx, settings.GetFileStoragePath()); err != nil {
		return nil, fmt.Errorf("failed to initialize data: %w", err)
	}

	AuthService := authService.NewAuthService(userRepo, jwtSettings)
	JWTService := jwtService.NewJWTService(jwtSettings)

	auditEventBus := audit.NewEventBus(settings.GetAuditFilePath(), settings.GetAuditURL(), logger)

	pingService := healthService.NewHealthCheckService(repoURLs)
	URLShortenerService := urlShortenerService.NewURLShortenerService(repoURLs, userURLsRepo, auditEventBus)
	URLExtractorService := urlExtractorService.NewLinkExtractorService(repoURLs, userURLsRepo, auditEventBus)
	URLDestructorService := urlDestructorService.NewURLDestructorService(repoURLs, userURLsRepo)
	StatsService := statsService.New(userRepo, repoURLs)

	URLExtractorHandler := urlExtractorHandler.NewExtractingFullLinkHandler(URLExtractorService)
	UserURLsHandler := userURLsHandler.NewExtractingUserURLsHandler(URLExtractorService, settings)
	URLShortenerHandler := urlShortenerHandler.NewCreatingShortLinksHandler(URLShortenerService, settings)
	URLShortenerAPIHandler := shortenAPI.NewCreatingShortURLsAPIHandler(URLShortenerService, settings)
	URLShortenerBatchAPIHandler := shortenBatchAPI.NewCreatingShortURLsByBatchAPIHandler(URLShortenerService, settings)
	URLDestructorAPIHandler := urlsDestructorAPIHandler.NewUsersURLsDestructorAPIHandler(URLDestructorService)
	HealthHandler := health.NewPingHandler(pingService)
	StatsHandler := statsHandler.New(StatsService, settings.GetTrustedSubnet())

	// Создаем и настраиваем gRPC сервер
	grpcServer := createGRPCServer(JWTService, AuthService, logger)
	grpcShortenerImpl := grpcImpl.NewRPCService(URLShortenerService, URLExtractorService, settings)

	// Регистрируем gRPC сервис
	pb.RegisterShortenerServiceServer(grpcServer, grpcShortenerImpl)

	// Включаем reflection для grpcurl и других инструментов (только для dev)
	if !settings.EnvSettings.Server.IsProd() {
		reflection.Register(grpcServer)
	}

	return &App{
		router:                    router,
		grpcServer:                grpcServer,
		shortLinksHandler:         URLShortenerHandler,
		shortLinksHandlerAPI:      URLShortenerAPIHandler,
		shortLinksBatchHandlerAPI: URLShortenerBatchAPIHandler,
		destructorAPIHandler:      URLDestructorAPIHandler,
		fullLinkHandler:           URLExtractorHandler,
		userURLsHandler:           UserURLsHandler,
		pingHandler:               HealthHandler,
		statsHandler:              StatsHandler,
		services: Services{
			auth:          AuthService,
			jwt:           JWTService,
			urlDestructor: URLDestructorService,
		},
		settings: settings,
		logger:   logger,
		dataBus: DataBus{
			auditEventBus: auditEventBus,
		},
	}, nil
}

// SetupCommonMiddlewares настраивает общие middleware для всех маршрутов.
// Добавляет middleware для request ID, логирования и сжатия ответов (gzip).
func (a *App) SetupCommonMiddlewares() {
	a.router.Use(middleware.RequestIDMiddleware(a.logger))
	a.router.Use(middleware.LoggerMiddleware(a.logger))
	a.router.Use(gzip.Middleware(a.logger))
}

// SetupRoutes настраивает маршруты приложения.
// Создает публичные и приватные группы маршрутов с соответствующими middleware.
// Настраивает маршруты для сокращения URL, получения URL, удаления URL и health check.
func (a *App) SetupRoutes() {
	internalMiddleware := middleware.NewInternalMiddleware(
		a.logger,
		a.settings.GetTrustedSubnet(),
	).InternalMiddlewareHandler()

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

	internalGroup := publicGroup.Group("/api/internal")
	internalGroup.Use(internalMiddleware)
	{
		internalGroup.GET("/stats", a.statsHandler.Handle)
	}

	privateGroup := a.router.Group("/")
	privateGroup.Use(anonNotAllowedMiddleware)
	{
		privateGroup.GET("/api/user/urls", a.userURLsHandler.Handle)
		privateGroup.DELETE("/api/user/urls", a.destructorAPIHandler.Handle)
	}

	a.router.GET("/:shortURL", a.fullLinkHandler.Handle)
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Добавляем pprof роуты для профилирования если не в проде
	if !a.settings.EnvSettings.Server.IsProd() {
		a.setupPprofRoutes()
	}
}

// Run запускает HTTP-сервер приложения на адресе, указанном в настройках.
// Возвращает ошибку, если сервер не может быть запущен.
func (a *App) Run() error {
	enableHTTPS := a.settings.IsHTTPSEnabled()
	if enableHTTPS {
		a.logger.Infof("Starting HTTPS server at %s", a.settings.GetServerAddress())
		err := http.ListenAndServeTLS(a.settings.GetServerAddress(), "cert.pem", "key.pem", a.router)
		if err != nil {
			return err
		}

		return nil
	}
	err := a.router.Run(a.settings.GetServerAddress())
	return err
}

func (a *App) RunGRPC() error {
	grpcAddr := a.settings.GetGRPCServerAddress()
	a.logger.Infof("Starting gRPC server at %s", grpcAddr)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	if err := a.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

func createGRPCServer(
	jwtService service.JWTService,
	authService service.AuthService,
	logger *zap.SugaredLogger,
) *grpc.Server {
	publicInterceptor := grpcMiddleware.JWTAuthInterceptor(
		jwtService,
		authService,
		true, // isAnonAllowed
		logger,
	)

	chain := grpc.ChainUnaryInterceptor(publicInterceptor)

	return grpc.NewServer(chain)
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

// Stop корректно останавливает приложение.
// Останавливает все фоновые сервисы (например, сервис удаления URL) и завершает работу приложения.
func (a *App) Stop() {
	a.logger.Info("Stopping application...")

	// Graceful shutdown gRPC сервера
	if a.grpcServer != nil {
		a.logger.Info("Stopping gRPC server...")
		a.grpcServer.GracefulStop()
	}

	if a.services.urlDestructor != nil {
		a.services.urlDestructor.Stop()
	}

	a.dataBus.auditEventBus.UnsubscribeAll()

	a.logger.Info("Application stopped")
}
