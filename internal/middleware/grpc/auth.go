package grpc

import (
	"context"
	"strings"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "Bearer "
)

func JWTAuthInterceptor(
	jwtService service.JWTService,
	authService service.AuthService,
	isAnonAllowed bool,
	logger *zap.SugaredLogger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Извлекаем метаданные из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			if isAnonAllowed {
				// Создаем анонимного пользователя
				return handler(createAnonymousUserContext(ctx, authService, logger), req)
			}
		}

		token, ok := extractTokenFromMetadata(md)
		if !ok {
			if isAnonAllowed {
				// Создаем анонимного пользователя
				return handler(createAnonymousUserContext(ctx, authService, logger), req)
			}
			logger.Warn("Authorization token not provided")
			return nil, status.Error(codes.Unauthenticated, "Authorization token not provided")
		}

		expired, err := jwtService.IsTokenExpired(ctx, token)
		if err != nil {
			logger.Errorw("Failed to check token expiration", "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if expired {
			logger.Debugw("JWT token expired")
			return nil, status.Error(codes.Unauthenticated, "token expired")
		}

		// Валидируем токен
		user, err := jwtService.ValidateToken(ctx, token)
		if err != nil {
			logger.Errorw("Failed to validate JWT token", "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Добавляем пользователя в контекст
		ctx = context.WithValue(ctx, middleware.JWTTokenContextKey, user)

		logger.Debugw("JWT authentication successful",
			"user_id", user.ID,
			"username", user.Name,
			"is_anonymous", user.IsAnonymous,
		)

		return handler(ctx, req)
	}
}

// createAnonymousUserContext создает контекст с анонимным пользователем
func createAnonymousUserContext(
	ctx context.Context,
	authService service.AuthService,
	logger *zap.SugaredLogger,
) context.Context {
	// Для gRPC получаем IP из peer информации
	md, _ := metadata.FromIncomingContext(ctx)
	clientIP := ""
	if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
		clientIP = ips[0]
	}
	userAgent := ""
	if uas := md.Get("user-agent"); len(uas) > 0 {
		userAgent = uas[0]
	}

	anonUser, err := authService.GetOrCreateAnonymousUser(ctx, clientIP, userAgent)
	if err != nil {
		logger.Errorw("Failed to get or create anonymous user", "error", err)
		return ctx
	}

	ctx = context.WithValue(ctx, middleware.JWTTokenContextKey, anonUser)
	return ctx
}

func extractTokenFromMetadata(md metadata.MD) (string, bool) {
	authHeaders := md.Get(authorizationHeader)
	if len(authHeaders) == 0 {
		return "", false
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", false
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", false
	}

	return token, true
}
