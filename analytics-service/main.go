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

	"analytics-service/internal/clickhouse"
	"analytics-service/internal/config"
	"analytics-service/internal/etl"

	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Load()

	// PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	defer sqlDB.Close()
	log.Println("connected to PostgreSQL")

	// ClickHouse
	chClient, err := clickhouse.NewClient(
		cfg.ClickHouseHost,
		cfg.ClickHouseDB,
		cfg.ClickHouseUser,
		cfg.ClickHousePass,
	)
	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}
	defer chClient.Close()
	log.Println("connected to ClickHouse")

	// Инициализация схем витрин
	if err := chClient.InitSchemas(); err != nil {
		log.Fatalf("failed to init ClickHouse schemas: %v", err)
	}
	log.Println("ClickHouse schemas ready")

	builders := etl.NewBuilders(db, chClient)

	// Полная перестройка при старте (в фоне, чтобы не задерживать /health)
	go func() {
		log.Println("running initial full ETL...")
		builders.RunFull()
		log.Println("initial ETL complete")
	}()

	// Ежедневный крон в 03:00 UTC
	c := cron.New()
	if _, err := c.AddFunc("0 3 * * *", func() {
		log.Println("running daily ETL...")
		builders.RunDaily()
		log.Println("daily ETL complete")
	}); err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// HTTP-сервер
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("/run-etl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		log.Println("/run-etl triggered manually")
		builders.RunFull()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","message":"ETL completed"}`)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("analytics-service listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down analytics-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("analytics-service stopped")
}
