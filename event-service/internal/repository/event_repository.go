package repository

import (
	"event-service/internal/models"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) CreateEvent(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) GetEventByID(id int) (*models.Event, error) {
	var event models.Event
	err := r.db.First(&event, id).Error
	return &event, err
}

func (r *EventRepository) ListEvents(creatorID *int, status string, categoryID *int) ([]models.Event, error) {
	var events []models.Event
	query := r.db.Order("created_at DESC")

	if creatorID != nil {
		query = query.Where("creator_id = ?", *creatorID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if categoryID != nil {
		query = query.Joins("JOIN event_categories ON event_categories.event_id = events.id").
			Where("event_categories.category_id = ?", *categoryID)
	}

	err := query.Find(&events).Error
	return events, err
}

func (r *EventRepository) UpdateEvent(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *EventRepository) DeleteEvent(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", id).Delete(&models.EventCategory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Event{}, id).Error
	})
}

func (r *EventRepository) AddEventCategories(eventID int, categoryIDs []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", eventID).Delete(&models.EventCategory{}).Error; err != nil {
			return err
		}

		for _, catID := range categoryIDs {
			if err := tx.Create(&models.EventCategory{
				EventID:    eventID,
				CategoryID: catID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *EventRepository) GetEventCategories(eventID int) ([]int, error) {
	var eventCategories []models.EventCategory
	err := r.db.Where("event_id = ?", eventID).Find(&eventCategories).Error
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]int, len(eventCategories))
	for i, ec := range eventCategories {
		categoryIDs[i] = ec.CategoryID
	}
	return categoryIDs, nil
}

func (r *EventRepository) ArchiveEvent(id int) error {
	return r.db.Model(&models.Event{}).Where("id = ?", id).Update("status", "archived").Error
}

func (r *EventRepository) PublishEvent(id int) error {
	return r.db.Model(&models.Event{}).Where("id = ?", id).Update("status", "published").Error
}
