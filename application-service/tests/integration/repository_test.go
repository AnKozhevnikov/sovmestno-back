//go:build integration

package integration

import (
	"application-service/internal/models"
	"application-service/internal/repository"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testDB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		fmt.Printf("failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	if err := migrateTestDB(testDB); err != nil {
		fmt.Printf("failed to migrate test database: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func migrateTestDB(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id          SERIAL PRIMARY KEY,
			creator_id  INT NOT NULL,
			title       VARCHAR(200) NOT NULL,
			is_active   BOOLEAN NOT NULL DEFAULT true,
			is_completed BOOLEAN NOT NULL DEFAULT false,
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS applications (
			id            SERIAL PRIMARY KEY,
			sender_id     INT NOT NULL,
			sender_type   VARCHAR(20) NOT NULL,
			receiver_id   INT NOT NULL,
			receiver_type VARCHAR(20) NOT NULL,
			event_id      INT NOT NULL,
			message       TEXT,
			status        VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE UNIQUE INDEX IF NOT EXISTS uq_application_pending
			ON applications (sender_id, receiver_id, event_id)
			WHERE status = 'pending';

		CREATE TABLE IF NOT EXISTS collaborations (
			id              SERIAL PRIMARY KEY,
			application_id  INT NOT NULL,
			event_id        INT NOT NULL,
			creator_user_id INT NOT NULL,
			venue_user_id   INT NOT NULL,
			status          VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error
}

func resetDB(t *testing.T) {
	t.Helper()
	if err := testDB.Exec("TRUNCATE applications, collaborations, events RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset db: %v", err)
	}
}

func seedEvent(t *testing.T, id, creatorID int) {
	t.Helper()
	if err := testDB.Exec(
		"INSERT INTO events (id, creator_id, title) VALUES (?, ?, ?)",
		id, creatorID, fmt.Sprintf("Event %d", id),
	).Error; err != nil {
		t.Fatalf("failed to seed event: %v", err)
	}
}

// ─── CreateApplication ────────────────────────────────────────────────────────

func TestIntegration_CreateApplication(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	app := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	if err := repo.CreateApplication(app); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
}

// ─── Partial unique index: нельзя дважды pending ──────────────────────────────

func TestIntegration_DuplicatePendingBlocked(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	app1 := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	if err := repo.CreateApplication(app1); err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	app2 := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	err := repo.CreateApplication(app2)
	if err != repository.ErrDuplicatePendingApplication {
		t.Errorf("expected ErrDuplicatePendingApplication, got %v", err)
	}
}

// ─── После reject дубль разрешён ─────────────────────────────────────────────

func TestIntegration_DuplicateAllowedAfterReject(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	app1 := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	if err := repo.CreateApplication(app1); err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	// Отклоняем
	app1.Status = "rejected"
	if err := repo.UpdateApplication(app1); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Теперь можно снова
	app2 := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	if err := repo.CreateApplication(app2); err != nil {
		t.Errorf("expected second pending to succeed after reject, got %v", err)
	}
}

// ─── AcceptApplicationTx: атомарность ────────────────────────────────────────

func TestIntegration_AcceptApplicationTx(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	app := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	if err := repo.CreateApplication(app); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	app.Status = "accepted"
	collab := &models.Collaboration{
		ApplicationID: app.ID, EventID: 1,
		CreatorUserID: 1, VenueUserID: 2,
		Status: "pending",
	}

	if err := repo.AcceptApplicationTx(app, collab); err != nil {
		t.Fatalf("accept tx failed: %v", err)
	}

	// Проверяем что event снят с каталога
	var isActive bool
	testDB.Raw("SELECT is_active FROM events WHERE id = 1").Scan(&isActive)
	if isActive {
		t.Error("expected event.is_active = false after accept")
	}

	// Проверяем что коллаборация создана
	if collab.ID == 0 {
		t.Error("expected collaboration ID to be set")
	}
}

// ─── CompleteCollaborationTx: event.is_completed ──────────────────────────────

func TestIntegration_CompleteCollaborationTx(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	app := &models.Application{
		SenderID: 1, SenderType: "creator",
		ReceiverID: 2, ReceiverType: "venue",
		EventID: 1, Status: "pending",
	}
	repo.CreateApplication(app)

	app.Status = "accepted"
	collab := &models.Collaboration{
		ApplicationID: app.ID, EventID: 1,
		CreatorUserID: 1, VenueUserID: 2, Status: "pending",
	}
	repo.AcceptApplicationTx(app, collab)

	if err := repo.CompleteCollaborationTx(collab.ID, 1); err != nil {
		t.Fatalf("complete tx failed: %v", err)
	}

	var isCompleted bool
	testDB.Raw("SELECT is_completed FROM events WHERE id = 1").Scan(&isCompleted)
	if !isCompleted {
		t.Error("expected event.is_completed = true after complete")
	}

	var collabStatus string
	testDB.Raw("SELECT status FROM collaborations WHERE id = ?", collab.ID).Scan(&collabStatus)
	if collabStatus != "completed" {
		t.Errorf("expected collab status completed, got %s", collabStatus)
	}
}

// ─── GetCompletedEventIDsByUserID ─────────────────────────────────────────────

func TestIntegration_GetCompletedEventIDsByUserID(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	seedEvent(t, 2, 1)
	repo := repository.NewApplicationRepository(testDB)

	// Коллаборация 1: completed, user=1 как creator
	testDB.Exec(`INSERT INTO collaborations (application_id, event_id, creator_user_id, venue_user_id, status)
		VALUES (1, 1, 1, 2, 'completed')`)
	// Коллаборация 2: completed, user=1 как venue
	testDB.Exec(`INSERT INTO collaborations (application_id, event_id, creator_user_id, venue_user_id, status)
		VALUES (2, 2, 3, 1, 'completed')`)
	// Коллаборация 3: pending — не должна войти
	testDB.Exec(`INSERT INTO collaborations (application_id, event_id, creator_user_id, venue_user_id, status)
		VALUES (3, 1, 1, 2, 'pending')`)

	ids, err := repo.GetCompletedEventIDsByUserID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 event IDs, got %d: %v", len(ids), ids)
	}
}

// ─── HasMirrorPendingApplication ─────────────────────────────────────────────

func TestIntegration_HasMirrorPendingApplication(t *testing.T) {
	resetDB(t)
	seedEvent(t, 1, 1)
	repo := repository.NewApplicationRepository(testDB)

	// venue отправил заявку creator'у
	testDB.Exec(`INSERT INTO applications (sender_id, sender_type, receiver_id, receiver_type, event_id, status)
		VALUES (2, 'venue', 1, 'creator', 1, 'pending')`)

	// creator хочет отправить заявку venue на тот же event — должен быть mirror
	hasMirror, err := repo.HasMirrorPendingApplication(1, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMirror {
		t.Error("expected mirror application to be detected")
	}

	// Для другого event — зеркала нет
	hasMirror, err = repo.HasMirrorPendingApplication(1, 2, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasMirror {
		t.Error("expected no mirror for different event")
	}
}
