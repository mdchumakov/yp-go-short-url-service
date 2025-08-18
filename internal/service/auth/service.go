package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

const maxUsernameLength = 256

func NewAuthService(userRepo repository.UserRepository, jwtSettings *config.JWTSettings) service.AuthService {
	return &authService{
		userRepo:    userRepo,
		jwtSettings: jwtSettings,
	}
}

type authService struct {
	userRepo    repository.UserRepository
	jwtSettings *config.JWTSettings
}

func (s *authService) CreateAnonymousUser(
	ctx context.Context,
	clientIP,
	userAgent string,
	expiresAt *time.Time,
) (*model.UserModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	username := s.GenerateAnonymousName(clientIP, userAgent)
	user, err := s.userRepo.CreateUser(ctx, username, "", expiresAt)
	if err != nil {
		logger.Errorw("Failed to create anonymous user", "username", username, "error", err, "request_id", requestID)
		return nil, err
	}
	return user, nil
}

func (s *authService) GetOrCreateAnonymousUser(ctx context.Context, clientIP, userAgent string) (*model.UserModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	username := s.GenerateAnonymousName(clientIP, userAgent)
	user, err := s.userRepo.GetUserByName(ctx, username)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		logger.Errorw("Failed to get anonymous user", "username", username, "error", err, "request_id", requestID)
		return nil, err
	}
	if user != nil {
		logger.Debugw("Anonymous user already exists", "username", username, "request_id", requestID)
		return user, nil
	}

	return s.CreateAnonymousUser(ctx, clientIP, userAgent, s.GenerateExpirationTime())
}

func (s *authService) GenerateAnonymousName(clientIP, userAgent string) string {
	// Создаем хеш из IP и User-Agent
	hash := sha256.Sum256([]byte(clientIP + userAgent))
	hashStr := hex.EncodeToString(hash[:])
	anonPrefix := "anon_"

	anonName := anonPrefix + hashStr
	if len(anonName) > maxUsernameLength {
		anonName = anonName[:maxUsernameLength]
	}
	return anonName
}

// CreateUser создает нового пользователя
func (s *authService) CreateUser(ctx context.Context, username, password string) (*model.UserModel, error) {
	panic("implement me")
}

// GetUserByID получает пользователя по ID
func (s *authService) GetUserByID(ctx context.Context, userID string) (*model.UserModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to get user by ID", "user_id", userID, "error", err, "request_id", requestID)
		return nil, err
	}

	return user, nil
}

// GetUserByName получает пользователя по имени
func (s *authService) GetUserByName(ctx context.Context, username string) (*model.UserModel, error) {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	user, err := s.userRepo.GetUserByName(ctx, username)
	if err != nil {
		logger.Errorw("Failed to get user by name", "username", username, "error", err, "request_id", requestID)
		return nil, err
	}

	return user, nil
}

func (s *authService) GenerateExpirationTime() *time.Time {
	expiration := time.Now().Add(s.jwtSettings.TokenDuration)
	return &expiration
}
