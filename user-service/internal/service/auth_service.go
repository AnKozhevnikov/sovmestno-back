package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo        *repository.UserRepository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config, redisClient *redis.Client) *AuthService {
	return &AuthService{
		repo:        repo,
		cfg:         cfg,
		redisClient: redisClient,
	}
}

type RegisterCreatorRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	WorkEmail   string `json:"work_email"`
	TgPersonal  string `json:"tg_personal_link"`
	TgChannel   string `json:"tg_channel_link"`
	VkLink      string `json:"vk_link"`
	TiktokLink  string `json:"tiktok_link"`
	YoutubeLink string `json:"youtube_link"`
	DzenLink    string `json:"dzen_link"`
}

type RegisterVenueRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=6"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	StreetAddress string `json:"street_address"`
	CityID        *int   `json:"city_id"`
	OpeningHours  string `json:"opening_hours"`
	Capacity      int    `json:"capacity"`
	Phone         string `json:"phone"`
	WorkEmail     string `json:"work_email"`
	TgPersonal    string `json:"tg_personal_link"`
	TgChannel     string `json:"tg_channel_link"`
	VkLink        string `json:"vk_link"`
	TiktokLink    string `json:"tiktok_link"`
	YoutubeLink   string `json:"youtube_link"`
	DzenLink      string `json:"dzen_link"`
	CategoryIDs  []int  `json:"category_ids"`
}

type RegisterAdminRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	AdminSecret string `json:"admin_secret" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"900"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         *models.User `json:"user"`
}

func (s *AuthService) RegisterCreator(req *RegisterCreatorRequest) (*AuthResponse, error) {
	// Проверяем, не существует ли пользователь
	existing, _ := s.repo.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "creator",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// Создаем профиль создателя
	creator := &models.Creator{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		Phone:       req.Phone,
		WorkEmail:   req.WorkEmail,
		TgPersonal:  req.TgPersonal,
		TgChannel:   req.TgChannel,
		VkLink:      req.VkLink,
		TiktokLink:  req.TiktokLink,
		YoutubeLink: req.YoutubeLink,
		DzenLink:    req.DzenLink,
	}
	if err := s.repo.CreateCreator(creator); err != nil {
		return nil, errors.New("failed to create creator profile: " + err.Error())
	}

	// Генерируем токены
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 минут
		User:         user,
	}, nil
}

func (s *AuthService) RegisterVenue(req *RegisterVenueRequest) (*AuthResponse, error) {
	// Проверяем, не существует ли пользователь
	existing, _ := s.repo.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "venue",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// Создаем профиль площадки
	venue := &models.Venue{
		UserID:        user.ID,
		Name:          req.Name,
		Description:   req.Description,
		StreetAddress: req.StreetAddress,
		CityID:        req.CityID,
		OpeningHours:  req.OpeningHours,
		Capacity:      req.Capacity,
		Phone:         req.Phone,
		WorkEmail:     req.WorkEmail,
		TgPersonal:    req.TgPersonal,
		TgChannel:     req.TgChannel,
		VkLink:        req.VkLink,
		TiktokLink:    req.TiktokLink,
		YoutubeLink:   req.YoutubeLink,
		DzenLink:      req.DzenLink,
	}
	if err := s.repo.CreateVenue(venue); err != nil {
		return nil, errors.New("failed to create venue profile: " + err.Error())
	}

	// Добавляем категории, если указаны
	if len(req.CategoryIDs) > 0 {
		if err := s.repo.AddVenueCategories(venue.ID, req.CategoryIDs); err != nil {
			return nil, errors.New("failed to add venue categories: " + err.Error())
		}
	}

	// Генерируем токены
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 минут
		User:         user,
	}, nil
}

func (s *AuthService) RegisterAdmin(req *RegisterAdminRequest) (*AuthResponse, error) {
	// Проверяем секретный ключ администратора
	if req.AdminSecret != s.cfg.AdminSecretKey {
		return nil, errors.New("invalid admin secret key")
	}

	// Проверяем, не существует ли пользователь
	existing, _ := s.repo.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя с ролью admin
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// Генерируем токены
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 минут
		User:         user,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Находим пользователя
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Генерируем токены
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 минут
		User:         user,
	}, nil
}

// generateAccessToken создает короткий access token (15 минут)
func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"type":    "access",
		"jti":     jti,
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // 15 минут
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// generateRefreshToken создает длинный refresh token (30 дней) и сохраняет в Redis
func (s *AuthService) generateRefreshToken(user *models.User) (string, error) {
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"type":    "refresh",
		"jti":     jti,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 дней
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	// Сохранить JTI в Redis whitelist
	ctx := context.Background()
	key := fmt.Sprintf("refresh:%s", jti)
	ttl := 30 * 24 * time.Hour

	err = s.redisClient.Set(ctx, key, user.ID, ttl).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save refresh token to Redis: %w", err)
	}

	return tokenString, nil
}

// RefreshAccessToken обновляет access token используя refresh token
func (s *AuthService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// 1. Распарсить refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// 2. Проверить тип токена
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return "", errors.New("token is not a refresh token")
	}

	// 3. Извлечь JTI и проверить в Redis whitelist
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return "", errors.New("token missing jti")
	}

	ctx := context.Background()
	key := fmt.Sprintf("refresh:%s", jti)

	exists := s.redisClient.Exists(ctx, key).Val()
	if exists == 0 {
		return "", errors.New("refresh token has been revoked or expired")
	}

	// 4. Извлечь user_id
	userIDFloat, _ := claims["user_id"].(float64)
	userID := int(userIDFloat)

	// 5. Получить пользователя из БД
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 6. Сгенерировать новый access token
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// Logout отзывает refresh token
func (s *AuthService) Logout(refreshTokenString string) error {
	// 1. Распарсить refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// 2. Извлечь JTI
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return errors.New("token missing jti")
	}

	// 3. Удалить из Redis whitelist
	ctx := context.Background()
	key := fmt.Sprintf("refresh:%s", jti)

	err = s.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// LogoutAll отзывает все refresh токены пользователя
func (s *AuthService) LogoutAll(userID int) error {
	ctx := context.Background()

	// Найти все ключи refresh токенов пользователя
	iter := s.redisClient.Scan(ctx, 0, "refresh:*", 0).Iterator()

	keysToDelete := []string{}

	for iter.Next(ctx) {
		key := iter.Val()

		// Проверить что это токен данного пользователя
		storedUserID, err := s.redisClient.Get(ctx, key).Int()
		if err != nil {
			continue
		}

		if storedUserID == userID {
			keysToDelete = append(keysToDelete, key)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// Удалить все найденные ключи
	if len(keysToDelete) > 0 {
		err := s.redisClient.Del(ctx, keysToDelete...).Err()
		if err != nil {
			return fmt.Errorf("failed to revoke all tokens: %w", err)
		}
	}

	return nil
}
