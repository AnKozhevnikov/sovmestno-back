//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
	"user-service/internal/config"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB  *gorm.DB
	testRDB *redis.Client
	testCfg *config.Config
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// PostgreSQL
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start postgres: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("failed to get pg conn string: %v\n", err)
		os.Exit(1)
	}

	testDB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		fmt.Printf("failed to connect to postgres: %v\n", err)
		os.Exit(1)
	}

	// Redis
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start redis: %v\n", err)
		os.Exit(1)
	}
	defer redisContainer.Terminate(ctx)

	redisAddr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		fmt.Printf("failed to get redis addr: %v\n", err)
		os.Exit(1)
	}

	testRDB = redis.NewClient(&redis.Options{Addr: redisAddr[8:]}) // strip "redis://"

	testCfg = &config.Config{
		JWTSecret:      "integration-test-secret-32bytes!",
		AdminSecretKey: "test-admin-secret",
	}

	if err := migrateTestDB(testDB); err != nil {
		fmt.Printf("migration failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func migrateTestDB(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            SERIAL PRIMARY KEY,
			email         VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role          VARCHAR(20)  NOT NULL,
			created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS images (
			id         UUID PRIMARY KEY,
			file_name  VARCHAR(255) NOT NULL,
			file_path  VARCHAR(500) NOT NULL,
			file_type  VARCHAR(100) NOT NULL,
			image_type VARCHAR(50)  NOT NULL,
			bucket_name VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS creators (
			id           SERIAL PRIMARY KEY,
			user_id      INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name         VARCHAR(100),
			description  TEXT,
			photo_id     UUID REFERENCES images(id) ON DELETE SET NULL,
			phone        VARCHAR(30),
			work_email   VARCHAR(255),
			tg_personal_link VARCHAR(255),
			tg_channel_link  VARCHAR(255),
			vk_link          VARCHAR(255),
			tiktok_link      VARCHAR(255),
			youtube_link     VARCHAR(255),
			dzen_link        VARCHAR(255),
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS venues (
			id             SERIAL PRIMARY KEY,
			user_id        INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name           VARCHAR(100),
			description    TEXT,
			street_address VARCHAR(500),
			city_id        INT,
			opening_hours  VARCHAR(200),
			capacity       INT,
			logo_id        UUID REFERENCES images(id) ON DELETE SET NULL,
			cover_photo_id UUID REFERENCES images(id) ON DELETE SET NULL,
			phone          VARCHAR(30),
			work_email     VARCHAR(255),
			tg_personal_link VARCHAR(255),
			tg_channel_link  VARCHAR(255),
			vk_link          VARCHAR(255),
			tiktok_link      VARCHAR(255),
			youtube_link     VARCHAR(255),
			dzen_link        VARCHAR(255),
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS newsletter_subscriptions (
			id                SERIAL PRIMARY KEY,
			email             VARCHAR(255) NOT NULL UNIQUE,
			unsubscribe_token UUID NOT NULL UNIQUE,
			subscribed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS creator_photos (
			id         SERIAL PRIMARY KEY,
			creator_id INT    NOT NULL REFERENCES creators(id) ON DELETE CASCADE,
			image_id   UUID   NOT NULL REFERENCES images(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS venue_photos (
			id       SERIAL PRIMARY KEY,
			venue_id INT  NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
			image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS creator_favorite_venues (
			creator_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			venue_user_id   INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (creator_user_id, venue_user_id)
		);
	`).Error
}

func resetDB(t *testing.T) {
	t.Helper()
	testDB.Exec("TRUNCATE creator_favorite_venues, newsletter_subscriptions, creators, venues, users RESTART IDENTITY CASCADE")
	testRDB.FlushAll(context.Background())
}

func newAuthSvc() *service.AuthService {
	repo := repository.NewUserRepository(testDB)
	return service.NewAuthService(repo, testCfg, testRDB)
}

// ─── RegisterCreator ──────────────────────────────────────────────────────────

func TestIntegration_RegisterCreator(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	resp, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email:    "creator@test.com",
		Password: "password123",
		Name:     "Test Creator",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected both tokens to be set")
	}
	if resp.User.Role != "creator" {
		t.Errorf("expected role creator, got %s", resp.User.Role)
	}

	// Проверяем что creator profile создан в БД
	var count int64
	testDB.Table("creators").Where("user_id = ?", resp.User.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 creator profile, got %d", count)
	}
}

func TestIntegration_RegisterCreator_DuplicateEmail(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "dup@test.com", Password: "password123", Name: "Test",
	})
	_, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "dup@test.com", Password: "password123", Name: "Test",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestIntegration_Login(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "user@test.com", Password: "password123", Name: "Test",
	})

	resp, err := svc.Login(&service.LoginRequest{
		Email: "user@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "user@test.com", Password: "password123", Name: "Test",
	})

	_, err := svc.Login(&service.LoginRequest{
		Email: "user@test.com", Password: "wrongpass",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestIntegration_RefreshToken(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	resp, _ := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "user@test.com", Password: "password123", Name: "Test",
	})

	newAccessToken, err := svc.RefreshAccessToken(resp.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newAccessToken == "" {
		t.Error("expected new access token")
	}
}

func TestIntegration_RefreshToken_AfterLogout(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	resp, _ := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "user@test.com", Password: "password123", Name: "Test",
	})

	// Логаутимся — refresh token должен стать недействительным
	svc.Logout(resp.RefreshToken)

	_, err := svc.RefreshAccessToken(resp.RefreshToken)
	if err == nil {
		t.Fatal("expected error after logout, got nil")
	}
}

// ─── Newsletter ───────────────────────────────────────────────────────────────

func TestIntegration_Newsletter(t *testing.T) {
	resetDB(t)
	repo := repository.NewUserRepository(testDB)
	svc := service.NewNewsletterService(repo)

	sub, err := svc.Subscribe("newsletter@test.com")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	if sub.UnsubscribeToken == "" {
		t.Error("expected unsubscribe token")
	}

	// Повторная подписка
	_, err = svc.Subscribe("newsletter@test.com")
	if err == nil {
		t.Fatal("expected error on duplicate subscribe")
	}

	// Отписка
	err = svc.UnsubscribeByToken(sub.UnsubscribeToken)
	if err != nil {
		t.Fatalf("unsubscribe failed: %v", err)
	}

	// После отписки токен недействителен
	err = svc.UnsubscribeByToken(sub.UnsubscribeToken)
	if err == nil {
		t.Fatal("expected error for used token")
	}
}

// ─── LogoutAll ────────────────────────────────────────────────────────────────

func TestIntegration_LogoutAll(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	// Регистрируемся дважды — получаем два разных refresh token
	resp1, _ := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "user@test.com", Password: "password123", Name: "Test",
	})
	resp2, err := svc.Login(&service.LoginRequest{
		Email: "user@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// LogoutAll должен инвалидировать оба токена
	svc.LogoutAll(resp1.User.ID)

	_, err = svc.RefreshAccessToken(resp1.RefreshToken)
	if err == nil {
		t.Error("expected error for first token after LogoutAll")
	}
	_, err = svc.RefreshAccessToken(resp2.RefreshToken)
	if err == nil {
		t.Error("expected error for second token after LogoutAll")
	}
}

// ─── CascadeDelete ────────────────────────────────────────────────────────────

func TestIntegration_CascadeDelete_UserDeletesCreator(t *testing.T) {
	resetDB(t)
	svc := newAuthSvc()

	resp, err := svc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "cascade@test.com", Password: "password123", Name: "ToDelete",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userID := resp.User.ID

	// Проверяем что creator создан
	var creatorCount int64
	testDB.Table("creators").Where("user_id = ?", userID).Count(&creatorCount)
	if creatorCount != 1 {
		t.Fatalf("expected 1 creator before delete, got %d", creatorCount)
	}

	// Удаляем пользователя
	testDB.Exec("DELETE FROM users WHERE id = ?", userID)

	// ON DELETE CASCADE должен удалить и creator
	testDB.Table("creators").Where("user_id = ?", userID).Count(&creatorCount)
	if creatorCount != 0 {
		t.Errorf("expected 0 creators after user delete (CASCADE), got %d", creatorCount)
	}
}

// ─── Favorites ────────────────────────────────────────────────────────────────

func TestIntegration_FavoriteVenues(t *testing.T) {
	resetDB(t)
	authSvc := newAuthSvc()

	// Создаём creator и venue
	creatorResp, _ := authSvc.RegisterCreator(&service.RegisterCreatorRequest{
		Email: "creator@test.com", Password: "pass1234", Name: "Creator",
	})
	venueResp, _ := authSvc.RegisterVenue(&service.RegisterVenueRequest{
		Email: "venue@test.com", Password: "pass1234", Name: "Venue",
	})

	repo := repository.NewUserRepository(testDB)
	favSvc := service.NewFavoritesService(repo)

	// Добавляем в избранное
	err := favSvc.AddFavoriteVenue(creatorResp.User.ID, venueResp.User.ID)
	if err != nil {
		t.Fatalf("add favorite failed: %v", err)
	}

	// Повторное добавление
	err = favSvc.AddFavoriteVenue(creatorResp.User.ID, venueResp.User.ID)
	if err == nil {
		t.Fatal("expected error on duplicate favorite")
	}

	// Список избранных
	venues, err := favSvc.ListFavoriteVenues(creatorResp.User.ID)
	if err != nil {
		t.Fatalf("list favorites failed: %v", err)
	}
	if len(venues) != 1 {
		t.Errorf("expected 1 favorite venue, got %d", len(venues))
	}

	// Удаляем из избранного
	err = favSvc.RemoveFavoriteVenue(creatorResp.User.ID, venueResp.User.ID)
	if err != nil {
		t.Fatalf("remove favorite failed: %v", err)
	}

	venues, _ = favSvc.ListFavoriteVenues(creatorResp.User.ID)
	if len(venues) != 0 {
		t.Errorf("expected 0 favorites after remove, got %d", len(venues))
	}
}
