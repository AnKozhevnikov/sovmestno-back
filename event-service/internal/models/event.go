package models

import "time"

type Event struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatorID    int       `gorm:"not null" json:"creator_id"`
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description,omitempty"`
	CoverPhotoID *int      `json:"cover_photo_id,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	IsCompleted  bool      `gorm:"default:false" json:"is_completed"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Categories   []int     `gorm:"-" json:"category_ids,omitempty"`
}

func (Event) TableName() string { return "events" }

type EventCategory struct {
	EventID    int `gorm:"primaryKey" json:"event_id"`
	CategoryID int `gorm:"primaryKey" json:"category_id"`
}

func (EventCategory) TableName() string { return "event_categories" }

// VenueFavoriteEvent - избранные мероприятия площадки
type VenueFavoriteEvent struct {
	VenueUserID int       `gorm:"primaryKey" json:"venue_user_id"`
	EventID     int       `gorm:"primaryKey" json:"event_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (VenueFavoriteEvent) TableName() string { return "venue_favorite_events" }
