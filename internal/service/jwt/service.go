package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

// Claims представляет структуру JWT claims
type Claims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	IsAnonymous bool   `json:"is_anonymous"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey     string
	tokenDuration time.Duration
	issuer        string
	algorithm     string
}

// NewJWTService создает новый экземпляр JWT сервиса
func NewJWTService(settings *config.JWTSettings) service.JWTService {
	return &jwtService{
		secretKey:     settings.SecretKey,
		tokenDuration: settings.TokenDuration,
		issuer:        settings.Issuer,
		algorithm:     settings.Algorithm,
	}
}

// GenerateToken генерирует JWT токен для пользователя
func (s *jwtService) GenerateToken(ctx context.Context, username string) (string, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	// Создаем claims для токена
	now := time.Now()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		logger.Errorw("Failed to sign JWT token", "username", username, "error", err, "request_id", requestID)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.Debugw("JWT token generated successfully", "username", username, "expires_at", claims.ExpiresAt, "request_id", requestID)
	return tokenString, nil
}

// GenerateTokenForUser генерирует JWT токен для конкретного пользователя
func (s *jwtService) GenerateTokenForUser(ctx context.Context, user *model.UserModel) (string, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	if user == nil {
		logger.Errorw("User cannot be nil", "request_id", requestID)
		return "", errors.New("user cannot be nil")
	}

	// Создаем claims для токена
	now := time.Now()
	claims := Claims{
		UserID:      user.ID,
		Username:    user.Name,
		IsAnonymous: user.IsAnonymous,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(user.ExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		logger.Errorw("Failed to sign JWT token", "user_id", user.ID, "username", user.Name, "error", err, "request_id", requestID)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	logger.Debugw("JWT token generated successfully",
		"user_id", user.ID,
		"username", user.Name,
		"is_anonymous", user.IsAnonymous,
		"expires_at", claims.ExpiresAt,
		"request_id", requestID,
	)
	return tokenString, nil
}

// ValidateToken валидирует JWT токен и возвращает информацию о пользователе
func (s *jwtService) ValidateToken(ctx context.Context, tokenString string) (*model.UserModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	if tokenString == "" {
		logger.Errorw("Token cannot be empty", "request_id", requestID)
		return nil, errors.New("token cannot be empty")
	}

	// Убираем префикс "Bearer" если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Errorw("Unexpected signing method", "method", token.Header["alg"], "request_id", requestID)
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		logger.Errorw("Failed to parse JWT token", "error", err, "request_id", requestID)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Проверяем валидность токена
	if !token.Valid {
		logger.Errorw("Invalid token", "error", err, "request_id", requestID)
		return nil, errors.New("invalid token")
	}

	// Извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Создаем модель пользователя
	user := &model.UserModel{
		ID:          claims.UserID,
		Name:        claims.Username,
		IsAnonymous: claims.IsAnonymous,
		CreatedAt:   claims.RegisteredClaims.IssuedAt.Time,
		UpdatedAt:   claims.RegisteredClaims.IssuedAt.Time,
	}

	logger.Debugw("JWT token validated successfully", "user_id", user.ID, "username", user.Name, "is_anonymous", user.IsAnonymous, "request_id", requestID)
	return user, nil
}

// GetUserIDFromToken извлекает ID пользователя из JWT токена
func (s *jwtService) GetUserIDFromToken(ctx context.Context, tokenString string) (string, error) {
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// RefreshToken обновляет JWT токен
func (s *jwtService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	// Валидируем текущий токен
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	// Генерируем новый токен
	return s.GenerateTokenForUser(ctx, user)
}

// IsTokenExpired проверяет, истек ли срок действия токена
func (s *jwtService) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	if tokenString == "" {
		return true, errors.New("token cannot be empty")
	}

	// Убираем префикс "Bearer" если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен без валидации подписи для проверки срока действия
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return true, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		logger.Errorw("Invalid token claims", "error", err, "request_id", requestID)
		return true, errors.New("invalid token claims")
	}

	// Проверяем срок действия
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		logger.Debugw("Token expired", "request_id", requestID)
		return true, nil
	}

	return false, nil
}

// GetTokenExpirationTime возвращает время истечения токена
func (s *jwtService) GetTokenExpirationTime(ctx context.Context, tokenString string) (*time.Time, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	if tokenString == "" {
		logger.Errorw("Token cannot be empty", "request_id", requestID)
		return nil, errors.New("token cannot be empty")
	}

	// Убираем префикс "Bearer " если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен без валидации подписи
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		logger.Errorw("Failed to parse JWT token", "error", err, "request_id", requestID)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		logger.Errorw("Invalid token claims", "error", err, "request_id", requestID)
		return nil, errors.New("invalid token claims")
	}

	if claims.ExpiresAt == nil {
		logger.Errorw("Token has no expiration time", "request_id", requestID)
		return nil, errors.New("token has no expiration time")
	}

	return &claims.ExpiresAt.Time, nil
}
