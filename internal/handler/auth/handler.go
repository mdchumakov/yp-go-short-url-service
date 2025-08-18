package auth

import (
	"net/http"
	"yp-go-short-url-service/internal/config"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	authService service.AuthService
	jwtService  service.JWTService
	jwtSettings *config.JWTSettings
	logger      *zap.SugaredLogger
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService service.AuthService, jwtService service.JWTService, jwtSettings *config.JWTSettings, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
		jwtSettings: jwtSettings,
		logger:      logger,
	}
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse представляет ответ с токеном
type AuthResponse struct {
	Token     string           `json:"token"`
	User      *model.UserModel `json:"user"`
	ExpiresIn int64            `json:"expires_in"`
}

// Register обрабатывает регистрацию нового пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorw("Invalid register request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Создаем нового пользователя
	user, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Errorw("Failed to create user", "username", req.Username, "error", err)
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
		return
	}

	// Генерируем JWT токен
	token, err := h.jwtService.GenerateTokenForUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Errorw("Failed to generate JWT token", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	// Получаем время истечения токена
	expTime, err := h.jwtService.GetTokenExpirationTime(c.Request.Context(), token)
	if err != nil {
		h.logger.Warnw("Failed to get token expiration time", "error", err)
	}

	var expiresIn int64
	if expTime != nil {
		expiresIn = expTime.Unix()
	}

	response := AuthResponse{
		Token:     token,
		User:      user,
		ExpiresIn: expiresIn,
	}

	// Устанавливаем куки
	middleware.SetJWTCookie(c, token, h.jwtSettings, h.jwtSettings.TokenDuration)

	h.logger.Infow("User registered successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusCreated, response)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorw("Invalid login request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Получаем пользователя по имени
	user, err := h.authService.GetUserByName(c.Request.Context(), req.Username)
	if err != nil {
		h.logger.Errorw("User not found", "username", req.Username, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Проверяем пароль (в реальном приложении нужно хешировать пароли)
	if user.Password != req.Password {
		h.logger.Errorw("Invalid password", "username", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Генерируем JWT токен
	token, err := h.jwtService.GenerateTokenForUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Errorw("Failed to generate JWT token", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	// Получаем время истечения токена
	expTime, err := h.jwtService.GetTokenExpirationTime(c.Request.Context(), token)
	if err != nil {
		h.logger.Warnw("Failed to get token expiration time", "error", err)
	}

	var expiresIn int64
	if expTime != nil {
		expiresIn = expTime.Unix()
	}

	response := AuthResponse{
		Token:     token,
		User:      user,
		ExpiresIn: expiresIn,
	}

	// Устанавливаем куки
	middleware.SetJWTCookie(c, token, h.jwtSettings, h.jwtSettings.TokenDuration)

	h.logger.Infow("User logged in successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusOK, response)
}

// RefreshToken обновляет JWT токен
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Получаем пользователя из контекста (должен быть аутентифицирован)
	user := middleware.GetJWTUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Errorw("User not found in context for token refresh")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Генерируем новый JWT токен
	token, err := h.jwtService.GenerateTokenForUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Errorw("Failed to generate new JWT token", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new authentication token",
		})
		return
	}

	// Получаем время истечения токена
	expTime, err := h.jwtService.GetTokenExpirationTime(c.Request.Context(), token)
	if err != nil {
		h.logger.Warnw("Failed to get token expiration time", "error", err)
	}

	var expiresIn int64
	if expTime != nil {
		expiresIn = expTime.Unix()
	}

	response := AuthResponse{
		Token:     token,
		User:      user,
		ExpiresIn: expiresIn,
	}

	// Устанавливаем куки
	middleware.SetJWTCookie(c, token, h.jwtSettings, h.jwtSettings.TokenDuration)

	h.logger.Infow("Token refreshed successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusOK, response)
}

// GetProfile возвращает профиль текущего пользователя
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Получаем пользователя из контекста
	user := middleware.GetJWTUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Errorw("User not found in context for profile request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Скрываем пароль в ответе
	user.Password = ""

	h.logger.Debugw("Profile retrieved successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// CreateAnonymousUser создает анонимного пользователя и возвращает токен
func (h *AuthHandler) CreateAnonymousUser(c *gin.Context) {
	// Создаем анонимного пользователя
	clientIP, userAgent := c.ClientIP(), c.Request.UserAgent()
	user, err := h.authService.CreateAnonymousUser(c.Request.Context(), clientIP, userAgent, nil)
	if err != nil {
		h.logger.Errorw("Failed to create anonymous user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create anonymous user",
		})
		return
	}

	// Генерируем JWT токен
	token, err := h.jwtService.GenerateTokenForUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Errorw("Failed to generate JWT token for anonymous user", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	// Получаем время истечения токена
	expTime, err := h.jwtService.GetTokenExpirationTime(c.Request.Context(), token)
	if err != nil {
		h.logger.Warnw("Failed to get token expiration time", "error", err)
	}

	var expiresIn int64
	if expTime != nil {
		expiresIn = expTime.Unix()
	}

	response := AuthResponse{
		Token:     token,
		User:      user,
		ExpiresIn: expiresIn,
	}

	// Устанавливаем куки
	middleware.SetJWTCookie(c, token, h.jwtSettings, h.jwtSettings.TokenDuration)

	h.logger.Infow("Anonymous user created successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusCreated, response)
}

// Logout обрабатывает выход пользователя из системы
func (h *AuthHandler) Logout(c *gin.Context) {
	// Получаем пользователя из контекста
	user := middleware.GetJWTUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Errorw("User not found in context for logout request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Удаляем куки
	middleware.ClearJWTCookie(c, h.jwtSettings)

	h.logger.Infow("User logged out successfully", "user_id", user.ID, "username", user.Name)
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
