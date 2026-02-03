package main

import (
	"event-service/internal/config"
	"event-service/internal/handlers"
	"event-service/internal/middleware"
	"event-service/internal/repository"
	"event-service/internal/service"
	"log"

	_ "event-service/docs" // Swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Event Service API
// @version 1.0
// @description API для управления мероприятиями
// @host localhost:8082
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg := config.Load()

	gin.SetMode(cfg.GinMode)

	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	eventRepo := repository.NewEventRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	eventService := service.NewEventService(eventRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	eventHandler := handlers.NewEventHandler(eventService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	events := r.Group("/events")
	events.Use(middleware.ExtractUserContext())
	{
		events.GET("", eventHandler.ListEvents)
		events.GET("/:id", eventHandler.GetEvent)
	}

	eventsCreator := r.Group("/events")
	eventsCreator.Use(middleware.ExtractUserContext(), middleware.RequireRole("creator"))
	{
		eventsCreator.POST("", eventHandler.CreateEvent)
		eventsCreator.PUT("/:id", eventHandler.UpdateEvent)
		eventsCreator.DELETE("/:id", eventHandler.DeleteEvent)
		eventsCreator.PATCH("/:id/archive", eventHandler.ArchiveEvent)
		eventsCreator.PATCH("/:id/publish", eventHandler.PublishEvent)
	}

	categories := r.Group("/categories")
	categories.Use(middleware.ExtractUserContext())
	{
		categories.GET("", categoryHandler.ListCategories)
		categories.GET("/:id", categoryHandler.GetCategory)
	}

	categoriesAdmin := r.Group("/categories")
	categoriesAdmin.Use(middleware.ExtractUserContext(), middleware.RequireRole("admin"))
	{
		categoriesAdmin.POST("", categoryHandler.CreateCategory)
		categoriesAdmin.PUT("/:id", categoryHandler.UpdateCategory)
		categoriesAdmin.DELETE("/:id", categoryHandler.DeleteCategory)
	}

	log.Printf("Event Service starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
