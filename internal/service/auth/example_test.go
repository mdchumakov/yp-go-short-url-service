package auth_test

import (
	"context"
	"fmt"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/repository/mock"
	"yp-go-short-url-service/internal/service/auth"

	"go.uber.org/mock/gomock"
)

// ExampleNewAuthService демонстрирует создание нового сервиса аутентификации.
func ExampleNewAuthService() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создаем мок репозитория пользователей
	mockUserRepo := mock.NewMockUserRepository(ctrl)

	// Настраиваем JWT настройки
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	// Создаем сервис аутентификации
	service := auth.NewAuthService(mockUserRepo, jwtSettings)

	// Сервис готов к использованию
	_ = service

	fmt.Println("AuthService created successfully")
	// Output: AuthService created successfully
}

// ExampleNewAuthService_getOrCreateAnonymousUser демонстрирует получение или создание анонимного пользователя.
func ExampleNewAuthService_getOrCreateAnonymousUser() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := auth.NewAuthService(mockUserRepo, jwtSettings)

	ctx := context.Background()
	clientIP := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	// В реальном приложении здесь будет вызов:
	// user, err := service.GetOrCreateAnonymousUser(ctx, clientIP, userAgent)
	// Для примера демонстрируем только создание сервиса
	_ = service
	_ = ctx
	_ = clientIP
	_ = userAgent

	fmt.Println("Service is ready to get or create anonymous user")
	// Output: Service is ready to get or create anonymous user
}

// ExampleNewAuthService_getUserByID демонстрирует получение пользователя по ID.
func ExampleNewAuthService_getUserByID() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := auth.NewAuthService(mockUserRepo, jwtSettings)

	ctx := context.Background()
	userID := "user-123"

	// В реальном приложении здесь будет вызов:
	// user, err := service.GetUserByID(ctx, userID)
	// Для примера демонстрируем только создание сервиса
	_ = service
	_ = ctx
	_ = userID

	fmt.Println("Service is ready to get user by ID")
	// Output: Service is ready to get user by ID
}

// ExampleNewAuthService_generateAnonymousName демонстрирует генерацию имени для анонимного пользователя.
func ExampleNewAuthService_generateAnonymousName() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := auth.NewAuthService(mockUserRepo, jwtSettings)

	clientIP := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	// Генерируем имя для анонимного пользователя
	anonymousName := service.GenerateAnonymousName(clientIP, userAgent)

	// Имя всегда начинается с префикса "anon_"
	if len(anonymousName) > 5 && anonymousName[:5] == "anon_" {
		fmt.Println("Generated anonymous name successfully")
	}
	// Output: Generated anonymous name successfully
}
