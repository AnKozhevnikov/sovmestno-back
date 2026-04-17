package repository

import (
	"event-service/internal/models"
	"fmt"

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

func (r *EventRepository) ListEvents(creatorID *int, categoryID *int, isActive *bool, isCompleted *bool, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	query := r.db.Order("created_at DESC")

	if creatorID != nil {
		query = query.Where("creator_id = ?", *creatorID)
	}

	if categoryID != nil {
		query = query.Joins("JOIN event_categories ON event_categories.event_id = events.id").
			Where("event_categories.category_id = ?", *categoryID)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if isCompleted != nil {
		query = query.Where("is_completed = ?", *isCompleted)
	}

	err := query.Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *EventRepository) PublishEvent(id int, creatorID int) error {
	result := r.db.Model(&models.Event{}).
		Where("id = ? AND creator_id = ?", id, creatorID).
		Update("is_active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found or access denied")
	}
	return nil
}

func (r *EventRepository) GetEventsByIDs(ids []int) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Where("id IN ?", ids).Find(&events).Error
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

// Favorites operations

func (r *EventRepository) AddVenueFavoriteEvent(venueUserID, eventID int) (bool, error) {
	fav := &models.VenueFavoriteEvent{
		VenueUserID: venueUserID,
		EventID:     eventID,
	}
	result := r.db.Where(fav).FirstOrCreate(fav)
	if result.Error != nil {
		return false, result.Error
	}
	alreadyExisted := result.RowsAffected == 0
	return alreadyExisted, nil
}

func (r *EventRepository) RemoveVenueFavoriteEvent(venueUserID, eventID int) error {
	result := r.db.Where("venue_user_id = ? AND event_id = ?", venueUserID, eventID).
		Delete(&models.VenueFavoriteEvent{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *EventRepository) ListVenueFavoriteEvents(venueUserID int) ([]models.Event, error) {
	var eventIDs []int
	if err := r.db.Model(&models.VenueFavoriteEvent{}).
		Where("venue_user_id = ?", venueUserID).
		Pluck("event_id", &eventIDs).Error; err != nil {
		return nil, err
	}
	if len(eventIDs) == 0 {
		return []models.Event{}, nil
	}
	var events []models.Event
	err := r.db.Where("id IN ?", eventIDs).Find(&events).Error
	return events, err
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

