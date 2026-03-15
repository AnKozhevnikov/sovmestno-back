package main

import (
	"application-service/internal/config"
	"application-service/internal/handlers"
	"application-service/internal/middleware"
	"application-service/internal/repository"
	"application-service/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "application-service/docs"
)

// @title Application Service API
// @version 1.0
// @description API для управления заявками на сотрудничество
// @host localhost:8083
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	applicationRepo := repository.NewApplicationRepository(db)
	applicationService := service.NewApplicationService(applicationRepo)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	collaborationHandler := handlers.NewCollaborationHandler(applicationService)

	r := gin.Default()

	r.Use(middleware.PrometheusMiddleware("application-service"))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	applications := r.Group("/applications")
	applications.Use(middleware.ExtractUserContext())
	{
		applications.POST("", applicationHandler.CreateApplication)
		applications.GET("", applicationHandler.ListApplications)
		applications.GET("/:id", applicationHandler.GetApplication)
		applications.PATCH("/:id/accept", applicationHandler.AcceptApplication)
		applications.PATCH("/:id/reject", applicationHandler.RejectApplication)
		applications.DELETE("/:id", applicationHandler.DeleteApplication)
	}

	collaborations := r.Group("/collaborations")
	collaborations.Use(middleware.ExtractUserContext())
	{
		collaborations.GET("", collaborationHandler.ListCollaborations)
		collaborations.GET("/partners", collaborationHandler.ListCollaborationPartners)
		collaborations.GET("/:id", collaborationHandler.GetCollaboration)
		collaborations.PATCH("/:id/complete", collaborationHandler.CompleteCollaboration)
		collaborations.PATCH("/:id/cancel", collaborationHandler.CancelCollaboration)
	}

	log.Printf("Application Service starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
