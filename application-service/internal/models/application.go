package models

import "time"

type Application struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID     int       `gorm:"not null" json:"sender_id"`
	SenderType   string    `gorm:"not null" json:"sender_type"`   // creator, venue
	ReceiverID   int       `gorm:"not null" json:"receiver_id"`
	ReceiverType string    `gorm:"not null" json:"receiver_type"` // creator, venue
	EventID      *int      `json:"event_id,omitempty"`
	Message      string    `json:"message,omitempty"`
	Status       string    `gorm:"default:pending" json:"status"` // pending, accepted, rejected
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Application) TableName() string { return "applications" }
