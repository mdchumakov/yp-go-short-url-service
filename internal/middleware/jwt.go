package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// jwtContextKey - тип для ключа контекста JWT
type jwtContextKey string

func (k jwtContextKey) String() string {
	return string(k)
}

const (
	JWTTokenContextKey  = jwtContextKey("jwt_user")
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

// JWTAuthMiddleware создает middleware для JWT аутентификации
func JWTAuthMiddleware(
	jwtService service.JWTService,
	authService service.AuthService,
	jwtSettings *config.JWTSettings,
	isAnonAllowed bool,
	logger *zap.SugaredLogger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		requestID := ExtractRequestID(ctx)

		// Получаем токен из заголовка Authorization
		token := extractTokenFromCookie(c, jwtSettings.CookieName)
		if token == "" {
			logger.Debugw("No JWT token found in request Cookie", "request_id", requestID)
			token = extractTokenFromHeader(c)
		}
		if token == "" {
			logger.Debugw("No JWT token found in request Headers", "request_id", requestID)
			clientIP, userAgent := c.ClientIP(), c.Request.UserAgent()
			if isAnonAllowed {
				anonUser, err := authService.GetOrCreateAnonymousUser(ctx, clientIP, userAgent)
				if err != nil {
					logger.Errorw("Failed to get or create anonymous user", "error", err, "request_id", requestID)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to authenticate user",
					})
					c.Abort()
					return
				}

				token, err = jwtService.GenerateTokenForUser(ctx, anonUser)
				if err != nil {
					logger.Errorw("Failed to generate token for anonymous user", "error", err, "request_id", requestID)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to generate authentication token",
					})
					c.Abort()
					return
				}

				// Устанавливаем куки для анонимного пользователя
				setJWTCookie(c, token, jwtSettings, logger, requestID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authentication required",
				})
				c.Abort()
				return
			}
		}

		// Проверяем, не истек ли токен
		expired, err := jwtService.IsTokenExpired(ctx, token)
		if err != nil {
			logger.Errorw("Failed to check token expiration", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		if expired {
			logger.Debugw("JWT token expired")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expired",
				"code":  "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		user, err := jwtService.ValidateToken(ctx, token)
		if err != nil {
			logger.Errorw("Failed to validate JWT token", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		setTokenInHeader(c, token)
		logger.Debugw("JWT token validated successfully", "user_id", user.ID, "request_id", requestID)

		// Добавляем пользователя в контекст
		ctx = context.WithValue(ctx, JWTTokenContextKey, user)
		c.Request = c.Request.WithContext(ctx)

		logger.Debugw(
			"JWT authentication successful",
			"user_id", user.ID,
			"username", user.Name,
			"is_anonymous", user.IsAnonymous,
			"request_id", requestID,
		)
		c.Next()

		setTokenInHeader(c, token)
	}
}

// setJWTCookie устанавливает JWT токен в куки
func setJWTCookie(c *gin.Context, token string, jwtSettings *config.JWTSettings, logger *zap.SugaredLogger, requestID string) {
	// Определяем домен для куки
	domain := jwtSettings.CookieDomain
	if domain == "" {
		domain = c.Request.Host
		if strings.Contains(domain, ":") {
			domain = strings.Split(domain, ":")[0]
		}
	}

	// Устанавливаем куки
	c.SetCookie(
		jwtSettings.CookieName,
		token,
		int(jwtSettings.TokenDuration.Seconds()),
		jwtSettings.CookiePath,
		domain,
		jwtSettings.CookieSecure,
		jwtSettings.CookieHTTPOnly,
	)

	if logger != nil {
		logger.Debugw("JWT cookie set successfully", "request_id", requestID)
	}
}

// SetJWTCookie устанавливает JWT токен в куки (публичная функция для использования в хендлерах)
func SetJWTCookie(c *gin.Context, token string, jwtSettings *config.JWTSettings, expiration time.Duration) {
	// Определяем домен для куки
	domain := jwtSettings.CookieDomain
	if domain == "" {
		domain = c.Request.Host
		if strings.Contains(domain, ":") {
			domain = strings.Split(domain, ":")[0]
		}
	}

	// Устанавливаем куки
	c.SetCookie(
		jwtSettings.CookieName,
		token,
		int(expiration.Seconds()),
		jwtSettings.CookiePath,
		domain,
		jwtSettings.CookieSecure,
		jwtSettings.CookieHTTPOnly,
	)
}

// ClearJWTCookie удаляет JWT куки
func ClearJWTCookie(c *gin.Context, jwtSettings *config.JWTSettings) {
	// Определяем домен для куки
	domain := jwtSettings.CookieDomain
	if domain == "" {
		domain = c.Request.Host
		if strings.Contains(domain, ":") {
			domain = strings.Split(domain, ":")[0]
		}
	}

	// Устанавливаем куки с отрицательным временем жизни для удаления
	c.SetCookie(
		jwtSettings.CookieName,
		"",
		-1,
		jwtSettings.CookiePath,
		domain,
		jwtSettings.CookieSecure,
		jwtSettings.CookieHTTPOnly,
	)
}

func extractTokenFromCookie(c *gin.Context, cookieName string) string {
	token, err := c.Cookie(cookieName)
	if err != nil {
		return ""
	}

	if token == "" {
		return ""
	}
	return token
}

// extractTokenFromHeader извлекает JWT токен из заголовка Authorization
func extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	// Проверяем формат "Bearer <token>"
	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	// Извлекаем токен (убираем префикс "Bearer ")
	token := strings.TrimPrefix(authHeader, BearerPrefix)
	if token == "" {
		return ""
	}

	return token
}

func setTokenInHeader(c *gin.Context, token string) {
	c.Header(AuthorizationHeader, BearerPrefix+token)
}

// GetJWTUserFromContext получает пользователя из JWT контекста
func GetJWTUserFromContext(ctx context.Context) *model.UserModel {
	if user, ok := ctx.Value(JWTTokenContextKey).(*model.UserModel); ok {
		return user
	}
	return nil
}
