package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/internal/config"
	"user-service/internal/handlers"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "user-service/docs" // Swagger generated docs
)

// @title           Совместно API - User Service
// @version         1.0
// @description     API для управления пользователями, профилями создателей и площадок
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@sovmestno.ru

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8081
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Настройка режима Gin
	if cfg.GinMode == "" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(cfg.GinMode)
	}

	// Подключаемся к БД
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Настройка connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Инициализация слоев приложения
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)

	// Инициализация ImageService
	imageService, err := service.NewImageService(userRepo, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize image service: %v", err)
	}

	userHandler := handlers.NewUserHandler(userService, imageService)

	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	// Настройка роутера
	r := gin.Default()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes (аутентификация)
	auth := r.Group("/auth")
	{
		auth.POST("/register/creator", authHandler.RegisterCreator)
		auth.POST("/register/venue", authHandler.RegisterVenue)
		auth.POST("/register/admin", authHandler.RegisterAdmin)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes (требуют аутентификации через X-User-ID header от gateway)
	users := r.Group("/users")
	users.Use(middleware.ExtractUserContext())
	{
		// Универсальный эндпоинт для получения профиля текущего пользователя
		users.GET("/me", userHandler.GetMe)

		// Профили создателей (creators) - создаются через /auth/register/creator
		users.GET("/creators/:user_id", userHandler.GetCreator)
		users.PUT("/creators/:user_id", userHandler.UpdateCreator)
		users.DELETE("/creators/:user_id", userHandler.DeleteCreator)

		// Профили площадок (venues) - создаются через /auth/register/venue
		users.GET("/venues/:user_id", userHandler.GetVenue)
		users.GET("/venues", userHandler.ListVenues)
		users.PUT("/venues/:user_id", userHandler.UpdateVenue)
		users.DELETE("/venues/:user_id", userHandler.DeleteVenue)

		// Загрузка изображений
		users.POST("/upload", userHandler.UploadImage)
		users.GET("/images/:id", userHandler.GetImage)
	}

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Port),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Starting user-service on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
