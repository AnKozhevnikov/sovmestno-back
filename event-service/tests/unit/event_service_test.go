package unit

import (
	"errors"
	"event-service/internal/models"
	"event-service/internal/service"
	"testing"
)

// ─── CreateEvent ──────────────────────────────────────────────────────────────

func TestCreateEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	event, err := svc.CreateEvent(&service.CreateEventRequest{Title: "My Event"}, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.CreatorID != 1 {
		t.Errorf("expected creatorID=1, got %d", event.CreatorID)
	}
	if event.Title != "My Event" {
		t.Errorf("unexpected title: %s", event.Title)
	}
}

func TestCreateEvent_WithCategories(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	event, err := svc.CreateEvent(
		&service.CreateEventRequest{Title: "Event", CategoryIDs: []int{1, 2, 3}},
		1,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(event.Categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(event.Categories))
	}
}

func TestCreateEvent_RepoError(t *testing.T) {
	repo := newMockEventRepo()
	repo.errCreate = errors.New("db error")
	svc := service.NewEventService(repo)

	_, err := svc.CreateEvent(&service.CreateEventRequest{Title: "Event"}, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── GetEventByID ─────────────────────────────────────────────────────────────

func TestGetEventByID_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Test", true, false)
	svc := service.NewEventService(repo)

	event, err := svc.GetEventByID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.ID != 1 {
		t.Errorf("expected ID=1, got %d", event.ID)
	}
}

func TestGetEventByID_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	_, err := svc.GetEventByID(999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── GetEventsByIDs ───────────────────────────────────────────────────────────

func TestGetEventsByIDs_Empty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	events, err := svc.GetEventsByIDs([]int{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty result, got %d events", len(events))
	}
}

func TestGetEventsByIDs_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Event 1", true, false)
	repo.events[2] = newEvent(2, 1, "Event 2", true, false)
	svc := service.NewEventService(repo)

	events, err := svc.GetEventsByIDs([]int{1, 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

// ─── UpdateEvent ─────────────────────────────────────────────────────────────

func TestUpdateEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Old Title", true, false)
	svc := service.NewEventService(repo)

	newTitle := "New Title"
	event, err := svc.UpdateEvent(1, &service.UpdateEventRequest{Title: &newTitle}, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %s", event.Title)
	}
}

func TestUpdateEvent_AccessDenied(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	svc := service.NewEventService(repo)

	title := "Hack"
	_, err := svc.UpdateEvent(1, &service.UpdateEventRequest{Title: &title}, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	title := "Title"
	_, err := svc.UpdateEvent(999, &service.UpdateEventRequest{Title: &title}, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── DeleteEvent ─────────────────────────────────────────────────────────────

func TestDeleteEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	svc := service.NewEventService(repo)

	if err := svc.DeleteEvent(1, 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.events) != 0 {
		t.Error("expected event to be deleted")
	}
}

func TestDeleteEvent_AccessDenied(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	svc := service.NewEventService(repo)

	err := svc.DeleteEvent(1, 99)
	if !errors.Is(err, service.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

// ─── ListEvents ───────────────────────────────────────────────────────────────

func TestListEvents_LimitNormalized(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	// limit=0 → нормализуется в 20
	_, err := svc.ListEvents(nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// limit=200 → нормализуется в 20
	_, err = svc.ListEvents(nil, nil, nil, nil, 200, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── FavoritesService ────────────────────────────────────────────────────────

func TestAddFavoriteEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	svc := service.NewFavoritesService(repo)

	if err := svc.AddFavoriteEvent(2, 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.favorites[2]) != 1 {
		t.Error("expected event to be added to favorites")
	}
}

func TestAddFavoriteEvent_EventNotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewFavoritesService(repo)

	err := svc.AddFavoriteEvent(2, 999)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestAddFavoriteEvent_AlreadyFavorited(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	repo.alreadyFaved = true
	svc := service.NewFavoritesService(repo)

	err := svc.AddFavoriteEvent(2, 1)
	if !errors.Is(err, service.ErrAlreadyFavorited) {
		t.Errorf("expected ErrAlreadyFavorited, got %v", err)
	}
}

func TestRemoveFavoriteEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Title", true, false)
	repo.favorites[2] = []int{1}
	svc := service.NewFavoritesService(repo)

	if err := svc.RemoveFavoriteEvent(2, 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.favorites[2]) != 0 {
		t.Error("expected favorite to be removed")
	}
}

func TestRemoveFavoriteEvent_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewFavoritesService(repo)

	err := svc.RemoveFavoriteEvent(2, 999)
	if !errors.Is(err, service.ErrFavoriteNotFound) {
		t.Errorf("expected ErrFavoriteNotFound, got %v", err)
	}
}

// ─── PublishEvent ─────────────────────────────────────────────────────────────

func TestPublishEvent_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Draft Event", false, false)
	svc := service.NewEventService(repo)

	if err := svc.PublishEvent(1, 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPublishEvent_WrongCreator(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 1, "Draft Event", false, false)
	svc := service.NewEventService(repo)

	err := svc.PublishEvent(1, 99)
	if err == nil {
		t.Fatal("expected error for wrong creator, got nil")
	}
}

// ─── ListFavoriteEvents ───────────────────────────────────────────────────────

func TestListFavoriteEvents_Success(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = newEvent(1, 10, "Event A", true, false)
	repo.events[2] = newEvent(2, 10, "Event B", true, false)
	repo.favorites[5] = []int{1, 2}
	svc := service.NewFavoritesService(repo)

	events, err := svc.ListFavoriteEvents(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 favorite events, got %d", len(events))
	}
}

func TestListFavoriteEvents_Empty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewFavoritesService(repo)

	events, err := svc.ListFavoriteEvents(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 favorite events, got %d", len(events))
	}
}

// ─── ListEvents with filters ──────────────────────────────────────────────────

func TestListEvents_FilterByCreator(t *testing.T) {
	repo := newMockEventRepo()
	creatorID := 42
	svc := service.NewEventService(repo)

	_, err := svc.ListEvents(&creatorID, nil, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListEvents_FilterByIsActive(t *testing.T) {
	repo := newMockEventRepo()
	isActive := true
	svc := service.NewEventService(repo)

	// categoryID=nil, isActive=true
	_, err := svc.ListEvents(nil, nil, &isActive, nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListEvents_FilterByIsCompleted(t *testing.T) {
	repo := newMockEventRepo()
	isCompleted := false
	svc := service.NewEventService(repo)

	_, err := svc.ListEvents(nil, nil, nil, &isCompleted, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── CategoryService ─────────────────────────────────────────────────────────

func TestCreateCategory_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := service.NewCategoryService(repo)

	cat, err := svc.CreateCategory(&service.CreateCategoryRequest{Name: "Music"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.Name != "Music" {
		t.Errorf("expected name 'Music', got %s", cat.Name)
	}
	if cat.ID == 0 {
		t.Error("expected ID to be set")
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	repo.categories[1] = &models.Category{ID: 1, Name: "Old"}
	svc := service.NewCategoryService(repo)

	cat, err := svc.UpdateCategory(1, &service.UpdateCategoryRequest{Name: "New"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.Name != "New" {
		t.Errorf("expected name 'New', got %s", cat.Name)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	repo.categories[1] = &models.Category{ID: 1, Name: "Music"}
	svc := service.NewCategoryService(repo)

	if err := svc.DeleteCategory(1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.categories) != 0 {
		t.Error("expected category to be deleted")
	}
}
