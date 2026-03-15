package models

import "time"

type Collaboration struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ApplicationID int       `gorm:"not null" json:"application_id"`
	EventID       int       `gorm:"not null" json:"event_id"`
	CreatorUserID int       `gorm:"not null" json:"creator_user_id"`
	VenueUserID   int       `gorm:"not null" json:"venue_user_id"`
	Status        string    `gorm:"default:pending" json:"status"` // pending, completed, cancelled
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Collaboration) TableName() string { return "collaborations" }
