package models

import "time"

type Event struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatorID    int       `gorm:"not null" json:"creator_id"`
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description,omitempty"`
	CoverPhotoID *int      `json:"cover_photo_id,omitempty"`
	Status       string    `gorm:"default:published" json:"status"` // published, archived
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Categories   []int     `gorm:"-" json:"category_ids,omitempty"` // Только для передачи данных
}

func (Event) TableName() string { return "events" }

type EventCategory struct {
	EventID    int `gorm:"primaryKey" json:"event_id"`
	CategoryID int `gorm:"primaryKey" json:"category_id"`
}

func (EventCategory) TableName() string { return "event_categories" }
