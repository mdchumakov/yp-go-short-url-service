package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock JWT Service
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(ctx context.Context, username string) (string, error) {
	args := m.Called(ctx, username)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateTokenForUser(ctx context.Context, user *model.UserModel) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(ctx context.Context, token string) (*model.UserModel, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockJWTService) GetUserIDFromToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) RefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) IsTokenExpired(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockJWTService) GetTokenExpirationTime(ctx context.Context, token string) (*time.Time, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

// Mock Auth Service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) CreateUser(ctx context.Context, username, password string) (*model.UserModel, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*model.UserModel, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockAuthService) GetUserByName(ctx context.Context, username string) (*model.UserModel, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockAuthService) CreateAnonymousUser(ctx context.Context, clientIP, userAgent string, expiresAt *time.Time) (*model.UserModel, error) {
	args := m.Called(ctx, clientIP, userAgent, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockAuthService) GetOrCreateAnonymousUser(ctx context.Context, clientIP, userAgent string) (*model.UserModel, error) {
	args := m.Called(ctx, clientIP, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserModel), args.Error(1)
}

func (m *MockAuthService) GenerateAnonymousName(clientIP, userAgent string) string {
	args := m.Called(clientIP, userAgent)
	return args.String(0)
}

func TestSetJWTCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовые настройки
	jwtSettings := &config.JWTSettings{
		CookieName:     "test_token",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: "lax",
		TokenDuration:  24 * time.Hour,
	}

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/test", nil)

	// Тестируем установку куки
	token := "test_jwt_token"
	logger := zap.NewNop().Sugar()
	setJWTCookie(c, token, jwtSettings, logger, "test_request_id")

	// Проверяем, что куки установлены
	result := w.Result()
	defer result.Body.Close()
	cookies := result.Cookies()
	assert.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "test_token", cookie.Name)
	assert.Equal(t, token, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, "localhost", cookie.Domain)
	assert.False(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, int(jwtSettings.TokenDuration.Seconds()), cookie.MaxAge)
}

func TestClearJWTCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовые настройки
	jwtSettings := &config.JWTSettings{
		CookieName:     "test_token",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: "lax",
	}

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/test", nil)

	// Тестируем удаление куки
	ClearJWTCookie(c, jwtSettings)

	// Проверяем, что куки установлены с отрицательным временем жизни
	result := w.Result()
	defer result.Body.Close()
	cookies := result.Cookies()
	assert.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "test_token", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
}

func TestExtractTokenFromCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)

	// Добавляем куки в запрос
	req.AddCookie(&http.Cookie{
		Name:  "test_token",
		Value: "test_jwt_token",
	})

	c.Request = req

	// Тестируем извлечение токена
	token := extractTokenFromCookie(c, "test_token")
	assert.Equal(t, "test_jwt_token", token)

	// Тестируем случай без куки
	token = extractTokenFromCookie(c, "non_existent_token")
	assert.Equal(t, "", token)
}

func TestJWTAuthMiddlewareWithCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовые настройки
	jwtSettings := &config.JWTSettings{
		CookieName:     "test_token",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: "lax",
		TokenDuration:  24 * time.Hour,
	}

	// Создаем моки
	mockJWTService := &MockJWTService{}
	mockAuthService := &MockAuthService{}

	logger := zap.NewNop().Sugar()

	// Создаем тестового пользователя
	testUser := &model.UserModel{
		ID:          "test_user_id",
		Name:        "test_user",
		Password:    "",
		IsAnonymous: false,
	}

	// Настраиваем ожидания для моков
	mockJWTService.On("IsTokenExpired", mock.Anything, "test_jwt_token").Return(false, nil)
	mockJWTService.On("ValidateToken", mock.Anything, "test_jwt_token").Return(testUser, nil)

	// Создаем middleware
	middleware := JWTAuthMiddleware(mockJWTService, mockAuthService, jwtSettings, false, logger)

	// Создаем тестовый контекст с куки
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "test_token",
		Value: "test_jwt_token",
	})
	c.Request = req

	// Добавляем request ID в контекст
	ctx := context.WithValue(c.Request.Context(), RequestIDKey, "test_request_id")
	c.Request = c.Request.WithContext(ctx)

	// Вызываем middleware
	middleware(c)

	// Проверяем, что пользователь добавлен в контекст
	user := GetJWTUserFromContext(c.Request.Context())
	assert.NotNil(t, user)
	assert.Equal(t, "test_user_id", user.ID)
	assert.Equal(t, "test_user", user.Name)

	// Проверяем, что моки были вызваны
	mockJWTService.AssertExpectations(t)
}
