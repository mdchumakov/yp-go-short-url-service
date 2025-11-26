package jwt_test

import (
	"context"
	"fmt"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service/jwt"
)

// ExampleNewJWTService демонстрирует создание нового JWT сервиса.
func ExampleNewJWTService() {
	// Настраиваем JWT настройки
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	// Создаем JWT сервис
	service := jwt.NewJWTService(jwtSettings)

	// Сервис готов к использованию
	_ = service

	fmt.Println("JWTService created successfully")
	// Output: JWTService created successfully
}

// ExampleNewJWTService_generateToken демонстрирует генерацию JWT токена для пользователя.
func ExampleNewJWTService_generateToken() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	username := "john_doe"

	// Генерируем токен для пользователя
	token, err := service.GenerateToken(ctx, username)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	if len(token) > 0 {
		fmt.Println("Token generated successfully")
	}
	// Output: Token generated successfully
}

// ExampleNewJWTService_generateTokenForUser демонстрирует генерацию JWT токена для конкретного пользователя.
func ExampleNewJWTService_generateTokenForUser() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()

	// Создаем модель пользователя
	user := &model.UserModel{
		ID:          "user-123",
		Name:        "john_doe",
		IsAnonymous: false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Генерируем токен для пользователя
	token, err := service.GenerateTokenForUser(ctx, user)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	fmt.Printf("Token generated for user %s\n", user.Name)
	_ = token
	// Output: Token generated for user john_doe
}

// ExampleNewJWTService_validateToken демонстрирует валидацию JWT токена.
func ExampleNewJWTService_validateToken() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	username := "john_doe"

	// Сначала генерируем токен
	token, err := service.GenerateToken(ctx, username)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	// Валидируем токен
	user, err := service.ValidateToken(ctx, token)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
		return
	}

	fmt.Printf("Token is valid for user: %s\n", user.Name)
	// Output: Token is valid for user: john_doe
}

// ExampleNewJWTService_getUserIDFromToken демонстрирует извлечение ID пользователя из токена.
func ExampleNewJWTService_getUserIDFromToken() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	user := &model.UserModel{
		ID:          "user-123",
		Name:        "john_doe",
		IsAnonymous: false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Генерируем токен для пользователя
	token, err := service.GenerateTokenForUser(ctx, user)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	// Извлекаем ID пользователя из токена
	userID, err := service.GetUserIDFromToken(ctx, token)
	if err != nil {
		fmt.Printf("Failed to get user ID: %v\n", err)
		return
	}

	fmt.Printf("User ID from token: %s\n", userID)
	// Output: User ID from token: user-123
}

// ExampleNewJWTService_refreshToken демонстрирует обновление JWT токена.
func ExampleNewJWTService_refreshToken() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	user := &model.UserModel{
		ID:          "user-123",
		Name:        "john_doe",
		IsAnonymous: false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Генерируем исходный токен
	oldToken, err := service.GenerateTokenForUser(ctx, user)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	// Обновляем токен
	newToken, err := service.RefreshToken(ctx, oldToken)
	if err != nil {
		fmt.Printf("Failed to refresh token: %v\n", err)
		return
	}

	if len(newToken) > 0 {
		fmt.Println("Token refreshed successfully")
	}
	// Output: Token refreshed successfully
}

// ExampleNewJWTService_isTokenExpired демонстрирует проверку истечения срока действия токена.
func ExampleNewJWTService_isTokenExpired() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 1 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	username := "john_doe"

	// Генерируем токен
	token, err := service.GenerateToken(ctx, username)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	// Проверяем, истек ли срок действия токена
	expired, err := service.IsTokenExpired(ctx, token)
	if err != nil {
		fmt.Printf("Failed to check token expiration: %v\n", err)
		return
	}

	if expired {
		fmt.Println("Token is expired")
	} else {
		fmt.Println("Token is valid")
	}
	// Output: Token is valid
}

// ExampleNewJWTService_getTokenExpirationTime демонстрирует получение времени истечения токена.
func ExampleNewJWTService_getTokenExpirationTime() {
	jwtSettings := &config.JWTSettings{
		SecretKey:     "your-secret-key-here-must-be-at-least-32-characters",
		TokenDuration: 24 * time.Hour,
		Issuer:        "yp-go-short-url-service",
		Algorithm:     "HS256",
	}

	service := jwt.NewJWTService(jwtSettings)

	ctx := context.Background()
	username := "john_doe"

	// Генерируем токен
	token, err := service.GenerateToken(ctx, username)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}

	// Получаем время истечения токена
	expirationTime, err := service.GetTokenExpirationTime(ctx, token)
	if err != nil {
		fmt.Printf("Failed to get expiration time: %v\n", err)
		return
	}

	if expirationTime != nil && expirationTime.After(time.Now()) {
		fmt.Println("Token expiration time retrieved successfully")
	}
	// Output: Token expiration time retrieved successfully
}
