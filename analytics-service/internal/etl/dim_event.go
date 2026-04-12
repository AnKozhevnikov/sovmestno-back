package etl

import (
	"analytics-service/internal/clickhouse"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type dimEventRow struct {
	EventID     int       `gorm:"column:event_id"`
	CreatorID   int       `gorm:"column:creator_id"`
	Title       string    `gorm:"column:title"`
	IsActive    bool      `gorm:"column:is_active"`
	IsCompleted bool      `gorm:"column:is_completed"`
	CreatedDate time.Time `gorm:"column:created_date"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type eventCategoryRow struct {
	EventID    int `gorm:"column:event_id"`
	CategoryID int `gorm:"column:category_id"`
}

func BuildDimEvent(db *gorm.DB, ch *clickhouse.Client) error {
	log.Println("[dim_event] extracting from PostgreSQL...")

	var events []dimEventRow
	if err := db.Raw(`
		SELECT id AS event_id, creator_id, title, is_active, is_completed,
		       DATE(created_at) AS created_date, updated_at
		FROM events
	`).Scan(&events).Error; err != nil {
		return fmt.Errorf("dim_event extract events: %w", err)
	}

	var catRows []eventCategoryRow
	if err := db.Raw(`SELECT event_id, category_id FROM event_categories`).Scan(&catRows).Error; err != nil {
		return fmt.Errorf("dim_event extract categories: %w", err)
	}

	// Строим map eventID → []categoryID
	catMap := make(map[int][]uint32, len(catRows))
	for _, c := range catRows {
		catMap[c.EventID] = append(catMap[c.EventID], uint32(c.CategoryID))
	}

	log.Printf("[dim_event] extracted %d events", len(events))

	ctx := context.Background()

	if err := ch.Exec(ctx, "TRUNCATE TABLE dim_event"); err != nil {
		return fmt.Errorf("dim_event truncate: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO dim_event")
	if err != nil {
		return fmt.Errorf("dim_event prepare batch: %w", err)
	}
	for _, e := range events {
		categoryIDs := catMap[e.EventID]
		if categoryIDs == nil {
			categoryIDs = []uint32{}
		}
		if err := batch.Append(
			uint32(e.EventID), uint32(e.CreatorID), e.Title,
			categoryIDs,
			boolToUint8(e.IsActive), boolToUint8(e.IsCompleted),
			e.CreatedDate, e.UpdatedAt,
		); err != nil {
			return fmt.Errorf("dim_event append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("dim_event send: %w", err)
	}

	log.Printf("[dim_event] loaded %d rows into ClickHouse", len(events))
	return nil
}
