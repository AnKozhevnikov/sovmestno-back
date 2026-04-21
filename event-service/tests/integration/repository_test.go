//go:build integration

package integration

import (
	"context"
	"event-service/internal/models"
	"event-service/internal/repository"
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
		CREATE TABLE IF NOT EXISTS categories (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS events (
			id            SERIAL PRIMARY KEY,
			creator_id    INT NOT NULL,
			title         VARCHAR(200) NOT NULL,
			description   TEXT,
			cover_photo_id UUID,
			is_active     BOOLEAN NOT NULL DEFAULT true,
			is_completed  BOOLEAN NOT NULL DEFAULT false,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS event_categories (
			event_id    INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			PRIMARY KEY (event_id, category_id)
		);

		CREATE TABLE IF NOT EXISTS venue_favorite_events (
			venue_user_id INT NOT NULL,
			event_id      INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (venue_user_id, event_id)
		);
	`).Error
}

func resetDB(t *testing.T) {
	t.Helper()
	if err := testDB.Exec("TRUNCATE venue_favorite_events, event_categories, events, categories RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset db: %v", err)
	}
}

// ─── CreateEvent / IsActive default ──────────────────────────────────────────

func TestIntegration_CreateEvent_IsActiveDefault(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	event := &models.Event{CreatorID: 1, Title: "Test Event"}
	if err := repo.CreateEvent(event); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fetched, err := repo.GetEventByID(event.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !fetched.IsActive {
		t.Error("expected is_active=true by default")
	}
	if fetched.IsCompleted {
		t.Error("expected is_completed=false by default")
	}
}

// ─── Categories ───────────────────────────────────────────────────────────────

func TestIntegration_AddEventCategories(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)
	catRepo := repository.NewCategoryRepository(testDB)

	cat1 := &models.Category{Name: "Music"}
	cat2 := &models.Category{Name: "Art"}
	catRepo.CreateCategory(cat1)
	catRepo.CreateCategory(cat2)

	event := &models.Event{CreatorID: 1, Title: "Event with cats"}
	repo.CreateEvent(event)

	if err := repo.AddEventCategories(event.ID, []int{cat1.ID, cat2.ID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ids, err := repo.GetEventCategories(event.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 categories, got %d", len(ids))
	}
}

func TestIntegration_AddEventCategories_Replaces(t *testing.T) {
	// Повторный вызов заменяет старые категории
	resetDB(t)
	repo := repository.NewEventRepository(testDB)
	catRepo := repository.NewCategoryRepository(testDB)

	cat1 := &models.Category{Name: "Music"}
	cat2 := &models.Category{Name: "Art"}
	catRepo.CreateCategory(cat1)
	catRepo.CreateCategory(cat2)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)

	repo.AddEventCategories(event.ID, []int{cat1.ID, cat2.ID})
	repo.AddEventCategories(event.ID, []int{cat1.ID}) // заменяем

	ids, _ := repo.GetEventCategories(event.ID)
	if len(ids) != 1 {
		t.Errorf("expected 1 category after replace, got %d", len(ids))
	}
}

// ─── DeleteEvent: CASCADE удаляет категории ───────────────────────────────────

func TestIntegration_DeleteEvent_CascadesCategories(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)
	catRepo := repository.NewCategoryRepository(testDB)

	cat := &models.Category{Name: "Music"}
	catRepo.CreateCategory(cat)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)
	repo.AddEventCategories(event.ID, []int{cat.ID})

	if err := repo.DeleteEvent(event.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	var count int64
	testDB.Model(&models.EventCategory{}).Where("event_id = ?", event.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected event_categories to be deleted, got %d rows", count)
	}
}

// ─── PublishEvent ─────────────────────────────────────────────────────────────

func TestIntegration_PublishEvent(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)
	// Снимаем с публикации вручную
	testDB.Model(&models.Event{}).Where("id = ?", event.ID).Update("is_active", false)

	if err := repo.PublishEvent(event.ID, 1); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	fetched, _ := repo.GetEventByID(event.ID)
	if !fetched.IsActive {
		t.Error("expected is_active=true after publish")
	}
}

func TestIntegration_PublishEvent_WrongCreator(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)

	err := repo.PublishEvent(event.ID, 99) // чужой creator
	if err == nil {
		t.Fatal("expected error for wrong creator, got nil")
	}
}

// ─── Favorites ────────────────────────────────────────────────────────────────

func TestIntegration_FavoriteEvent_AddRemove(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)

	alreadyExisted, err := repo.AddVenueFavoriteEvent(2, event.ID)
	if err != nil {
		t.Fatalf("add favorite failed: %v", err)
	}
	if alreadyExisted {
		t.Error("expected alreadyExisted=false on first add")
	}

	// Повторное добавление
	alreadyExisted, err = repo.AddVenueFavoriteEvent(2, event.ID)
	if err != nil {
		t.Fatalf("second add failed: %v", err)
	}
	if !alreadyExisted {
		t.Error("expected alreadyExisted=true on second add")
	}

	// Удаление
	if err := repo.RemoveVenueFavoriteEvent(2, event.ID); err != nil {
		t.Fatalf("remove favorite failed: %v", err)
	}

	events, _ := repo.ListVenueFavoriteEvents(2)
	if len(events) != 0 {
		t.Errorf("expected empty favorites, got %d", len(events))
	}
}

func TestIntegration_FavoriteEvent_CascadeDelete(t *testing.T) {
	// При удалении event избранное удаляется каскадно
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	event := &models.Event{CreatorID: 1, Title: "Event"}
	repo.CreateEvent(event)
	repo.AddVenueFavoriteEvent(2, event.ID)

	repo.DeleteEvent(event.ID)

	var count int64
	testDB.Model(&models.VenueFavoriteEvent{}).Where("event_id = ?", event.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected favorites to be deleted via cascade, got %d", count)
	}
}

// ─── GetEventsByIDs ───────────────────────────────────────────────────────────

func TestIntegration_GetEventsByIDs(t *testing.T) {
	resetDB(t)
	repo := repository.NewEventRepository(testDB)

	e1 := &models.Event{CreatorID: 1, Title: "E1"}
	e2 := &models.Event{CreatorID: 1, Title: "E2"}
	repo.CreateEvent(e1)
	repo.CreateEvent(e2)

	events, err := repo.GetEventsByIDs([]int{e1.ID, e2.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}
