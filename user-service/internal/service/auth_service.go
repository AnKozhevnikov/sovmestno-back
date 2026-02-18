package service

import (
	"errors"
	"time"
	"user-service/internal/config"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo: repo,
		cfg:  cfg,
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
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Address      string `json:"address"`
	OpeningHours string `json:"opening_hours"`
	Capacity     int    `json:"capacity"`
	Phone        string `json:"phone"`
	WorkEmail    string `json:"work_email"`
	TgPersonal   string `json:"tg_personal_link"`
	TgChannel    string `json:"tg_channel_link"`
	VkLink       string `json:"vk_link"`
	TiktokLink   string `json:"tiktok_link"`
	YoutubeLink  string `json:"youtube_link"`
	DzenLink     string `json:"dzen_link"`
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

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
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

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
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
		UserID:       user.ID,
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		OpeningHours: req.OpeningHours,
		Capacity:     req.Capacity,
		Phone:        req.Phone,
		WorkEmail:    req.WorkEmail,
		TgPersonal:   req.TgPersonal,
		TgChannel:    req.TgChannel,
		VkLink:       req.VkLink,
		TiktokLink:   req.TiktokLink,
		YoutubeLink:  req.YoutubeLink,
		DzenLink:     req.DzenLink,
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

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
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

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
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

	// Генерируем токен
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Токен действует 24 часа
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
