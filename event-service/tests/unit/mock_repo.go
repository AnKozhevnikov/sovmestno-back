package unit

import (
	"errors"
	"event-service/internal/models"

	"gorm.io/gorm"
)

var errNotFound = errors.New("not found")

type mockEventRepo struct {
	events         map[int]*models.Event
	categories     map[int][]int // eventID -> []categoryID
	favorites      map[int][]int // venueUserID -> []eventID
	nextID         int
	errCreate      error
	errGetByID     error
	errUpdate      error
	errDelete      error
	errPublish     error
	errAddFav      error
	alreadyFaved   bool
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events:     make(map[int]*models.Event),
		categories: make(map[int][]int),
		favorites:  make(map[int][]int),
		nextID:     1,
	}
}

func (m *mockEventRepo) CreateEvent(event *models.Event) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	event.ID = m.nextID
	m.nextID++
	cp := *event
	m.events[event.ID] = &cp
	return nil
}

func (m *mockEventRepo) GetEventByID(id int) (*models.Event, error) {
	if m.errGetByID != nil {
		return nil, m.errGetByID
	}
	e, ok := m.events[id]
	if !ok {
		return nil, errNotFound
	}
	cp := *e
	return &cp, nil
}

func (m *mockEventRepo) GetEventsByIDs(ids []int) ([]models.Event, error) {
	var result []models.Event
	for _, id := range ids {
		if e, ok := m.events[id]; ok {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockEventRepo) ListEvents(creatorID *int, categoryID *int, isActive *bool, isCompleted *bool, limit, offset int) ([]models.Event, error) {
	var result []models.Event
	for _, e := range m.events {
		if creatorID != nil && e.CreatorID != *creatorID {
			continue
		}
		if isActive != nil && e.IsActive != *isActive {
			continue
		}
		if isCompleted != nil && e.IsCompleted != *isCompleted {
			continue
		}
		result = append(result, *e)
	}
	return result, nil
}

func (m *mockEventRepo) UpdateEvent(event *models.Event) error {
	if m.errUpdate != nil {
		return m.errUpdate
	}
	cp := *event
	m.events[event.ID] = &cp
	return nil
}

func (m *mockEventRepo) DeleteEvent(id int) error {
	if m.errDelete != nil {
		return m.errDelete
	}
	delete(m.events, id)
	delete(m.categories, id)
	return nil
}

func (m *mockEventRepo) PublishEvent(id int, creatorID int) error {
	if m.errPublish != nil {
		return m.errPublish
	}
	e, ok := m.events[id]
	if !ok || e.CreatorID != creatorID {
		return errors.New("event not found or access denied")
	}
	e.IsActive = true
	return nil
}

func (m *mockEventRepo) AddEventCategories(eventID int, categoryIDs []int) error {
	m.categories[eventID] = categoryIDs
	return nil
}

func (m *mockEventRepo) GetEventCategories(eventID int) ([]int, error) {
	return m.categories[eventID], nil
}

func (m *mockEventRepo) AddVenueFavoriteEvent(venueUserID, eventID int) (bool, error) {
	if m.errAddFav != nil {
		return false, m.errAddFav
	}
	if m.alreadyFaved {
		return true, nil
	}
	m.favorites[venueUserID] = append(m.favorites[venueUserID], eventID)
	return false, nil
}

func (m *mockEventRepo) RemoveVenueFavoriteEvent(venueUserID, eventID int) error {
	list := m.favorites[venueUserID]
	for i, id := range list {
		if id == eventID {
			m.favorites[venueUserID] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *mockEventRepo) ListVenueFavoriteEvents(venueUserID int) ([]models.Event, error) {
	var result []models.Event
	for _, id := range m.favorites[venueUserID] {
		if e, ok := m.events[id]; ok {
			result = append(result, *e)
		}
	}
	return result, nil
}

// CategoryRepository mock

type mockCategoryRepo struct {
	categories map[int]*models.Category
	nextID     int
	errCreate  error
	errGet     error
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		categories: make(map[int]*models.Category),
		nextID:     1,
	}
}

func (m *mockCategoryRepo) CreateCategory(category *models.Category) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	category.ID = m.nextID
	m.nextID++
	cp := *category
	m.categories[category.ID] = &cp
	return nil
}

func (m *mockCategoryRepo) GetCategoryByID(id int) (*models.Category, error) {
	if m.errGet != nil {
		return nil, m.errGet
	}
	c, ok := m.categories[id]
	if !ok {
		return nil, errNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockCategoryRepo) ListCategories() ([]models.Category, error) {
	var result []models.Category
	for _, c := range m.categories {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockCategoryRepo) UpdateCategory(category *models.Category) error {
	cp := *category
	m.categories[category.ID] = &cp
	return nil
}

func (m *mockCategoryRepo) DeleteCategory(id int) error {
	delete(m.categories, id)
	return nil
}

// helpers

func newEvent(id, creatorID int, title string, isActive, isCompleted bool) *models.Event {
	return &models.Event{
		ID:          id,
		CreatorID:   creatorID,
		Title:       title,
		IsActive:    isActive,
		IsCompleted: isCompleted,
	}
}
