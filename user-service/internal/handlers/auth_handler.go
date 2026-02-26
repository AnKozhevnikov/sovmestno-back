package handlers

import (
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterCreator godoc
// @Summary      Регистрация создателя мероприятий
// @Description  Создает аккаунт и профиль создателя мероприятий, возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterCreatorRequest true "Данные регистрации создателя"
// @Success      201 {object} service.AuthResponse "Успешная регистрация"
// @Failure      400 {object} map[string]string "Ошибка валидации или email уже существует"
// @Router       /auth/register/creator [post]
func (h *AuthHandler) RegisterCreator(c *gin.Context) {
	var req service.RegisterCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RegisterCreator(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, resp)
}

// RegisterVenue godoc
// @Summary      Регистрация площадки
// @Description  Создает аккаунт и профиль площадки, возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterVenueRequest true "Данные регистрации площадки"
// @Success      201 {object} service.AuthResponse "Успешная регистрация"
// @Failure      400 {object} map[string]string "Ошибка валидации или email уже существует"
// @Router       /auth/register/venue [post]
func (h *AuthHandler) RegisterVenue(c *gin.Context) {
	var req service.RegisterVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RegisterVenue(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, resp)
}

// RegisterAdmin godoc
// @Summary      Регистрация администратора
// @Description  Создает аккаунт администратора (требует секретный ключ), возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterAdminRequest true "Данные регистрации админа"
// @Success      201 {object} service.AuthResponse "Успешная регистрация"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      403 {object} map[string]string "Неверный секретный ключ"
// @Router       /auth/register/admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	var req service.RegisterAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RegisterAdmin(&req)
	if err != nil {
		if err.Error() == "invalid admin secret key" {
			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, resp)
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентифицирует пользователя и возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "Данные для входа"
// @Success      200 {object} service.AuthResponse "Успешный вход"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Неверный email или пароль"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

// RefreshToken godoc
// @Summary      Обновление access token
// @Description  Получает новый access token используя refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} service.RefreshTokenResponse "Новый access token"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Невалидный refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input service.RefreshTokenRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "refresh_token is required"})
		return
	}

	newAccessToken, err := h.authService.RefreshAccessToken(input.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	resp := service.RefreshTokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   900,
	}

	c.JSON(200, resp)
}

// Logout godoc
// @Summary      Выход из системы
// @Description  Отзывает refresh token (logout с текущего устройства)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.LogoutRequest true "Refresh token"
// @Success      200 {object} map[string]string "Успешный выход"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var input service.LogoutRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "refresh_token is required"})
		return
	}

	err := h.authService.Logout(input.RefreshToken)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully logged out"})
}

// LogoutAll godoc
// @Summary      Выход со всех устройств
// @Description  Отзывает все refresh токены пользователя (logout со всех устройств)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Успешный выход со всех устройств"
// @Failure      401 {object} map[string]string "Неавторизован"
// @Failure      500 {object} map[string]string "Ошибка сервера"
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Получить user_id из контекста (установлен middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.authService.LogoutAll(userID.(int))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to logout from all devices"})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully logged out from all devices"})
}
