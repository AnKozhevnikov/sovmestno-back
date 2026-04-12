package handlers

import (
	"errors"
	"net/http"
	"user-service/internal/apperror"
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
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /auth/register/creator [post]
func (h *AuthHandler) RegisterCreator(c *gin.Context) {
	var req service.RegisterCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.authService.RegisterCreator(&req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, apperror.One("EMAIL_ALREADY_EXISTS", "User with this email already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create account"))
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RegisterVenue godoc
// @Summary      Регистрация площадки
// @Description  Создает аккаунт и профиль площадки, возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterVenueRequest true "Данные регистрации площадки"
// @Success      201 {object} service.AuthResponse "Успешная регистрация"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /auth/register/venue [post]
func (h *AuthHandler) RegisterVenue(c *gin.Context) {
	var req service.RegisterVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.authService.RegisterVenue(&req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, apperror.One("EMAIL_ALREADY_EXISTS", "User with this email already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create account"))
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RegisterAdmin godoc
// @Summary      Регистрация администратора
// @Description  Создает аккаунт администратора (требует секретный ключ), возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterAdminRequest true "Данные регистрации админа"
// @Success      201 {object} service.AuthResponse "Успешная регистрация"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /auth/register/admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	var req service.RegisterAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.authService.RegisterAdmin(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAdminSecret) {
			c.JSON(http.StatusForbidden, apperror.One("INVALID_ADMIN_SECRET", "Invalid admin secret key"))
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, apperror.One("EMAIL_ALREADY_EXISTS", "User with this email already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to create account"))
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентифицирует пользователя и возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "Данные для входа"
// @Success      200 {object} service.AuthResponse "Успешный вход"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if resp, ok := apperror.FromValidation(err); ok {
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusBadRequest, apperror.One("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, apperror.One("INVALID_CREDENTIALS", "Invalid email or password"))
			return
		}
		c.JSON(http.StatusUnauthorized, apperror.One("INVALID_CREDENTIALS", "Invalid email or password"))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary      Обновление access token
// @Description  Получает новый access token используя refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} service.RefreshTokenResponse "Новый access token"
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input service.RefreshTokenRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("FIELD_REQUIRED", "refresh_token is required"))
		return
	}

	newAccessToken, err := h.authService.RefreshAccessToken(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apperror.One("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token"))
		return
	}

	resp := service.RefreshTokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   900,
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      Выход из системы
// @Description  Отзывает refresh token (logout с текущего устройства)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.LogoutRequest true "Refresh token"
// @Success      200 {object} map[string]string "Успешный выход"
// @Failure      400 {object} apperror.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var input service.LogoutRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("FIELD_REQUIRED", "refresh_token is required"))
		return
	}

	if err := h.authService.Logout(input.RefreshToken); err != nil {
		c.JSON(http.StatusBadRequest, apperror.One("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// LogoutAll godoc
// @Summary      Выход со всех устройств
// @Description  Отзывает все refresh токены пользователя (logout со всех устройств)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Успешный выход со всех устройств"
// @Failure      401 {object} apperror.ErrorResponse
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, apperror.One("UNAUTHORIZED", "Unauthorized"))
		return
	}

	if err := h.authService.LogoutAll(userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, apperror.One("INTERNAL_ERROR", "Failed to logout from all devices"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out from all devices"})
}
