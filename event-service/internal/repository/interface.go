package repository

import "event-service/internal/models"

type EventRepositoryInterface interface {
	CreateEvent(event *models.Event) error
	GetEventByID(id int) (*models.Event, error)
	GetEventsByIDs(ids []int) ([]models.Event, error)
	ListEvents(creatorID *int, categoryID *int, isActive *bool, isCompleted *bool, limit, offset int) ([]models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id int) error
	PublishEvent(id int, creatorID int) error
	AddEventCategories(eventID int, categoryIDs []int) error
	GetEventCategories(eventID int) ([]int, error)
	AddVenueFavoriteEvent(venueUserID, eventID int) (bool, error)
	RemoveVenueFavoriteEvent(venueUserID, eventID int) error
	ListVenueFavoriteEvents(venueUserID int) ([]models.Event, error)
}

type CategoryRepositoryInterface interface {
	CreateCategory(category *models.Category) error
	GetCategoryByID(id int) (*models.Category, error)
	ListCategories() ([]models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id int) error
}
