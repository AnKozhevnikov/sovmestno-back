package etl

import (
	"analytics-service/internal/clickhouse"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type factCollaborationRow struct {
	CollaborationID int       `gorm:"column:collaboration_id"`
	Date            time.Time `gorm:"column:date"`
	ApplicationID   int       `gorm:"column:application_id"`
	EventID         int       `gorm:"column:event_id"`
	CreatorID       int       `gorm:"column:creator_id"`
	VenueID         int       `gorm:"column:venue_id"`
	Status          string    `gorm:"column:status"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

const factCollaborationsQuery = `
SELECT id AS collaboration_id, DATE(created_at) AS date,
       application_id, event_id, creator_user_id AS creator_id, venue_user_id AS venue_id,
       status, created_at, updated_at
FROM collaborations
`

const factCollaborationsIncrementalQuery = `
SELECT id AS collaboration_id, DATE(created_at) AS date,
       application_id, event_id, creator_user_id AS creator_id, venue_user_id AS venue_id,
       status, created_at, updated_at
FROM collaborations
WHERE updated_at >= ?
`

// BuildFactCollaborationsAll полная перестройка.
func BuildFactCollaborationsAll(db *gorm.DB, ch *clickhouse.Client) error {
	return buildFactCollaborations(db, ch, factCollaborationsQuery)
}

// BuildFactCollaborationsIncremental обновляет строки, изменившиеся с указанной даты.
func BuildFactCollaborationsIncremental(db *gorm.DB, ch *clickhouse.Client, since time.Time) error {
	return buildFactCollaborations(db, ch, factCollaborationsIncrementalQuery, since.Format("2006-01-02"))
}

func buildFactCollaborations(db *gorm.DB, ch *clickhouse.Client, query string, args ...interface{}) error {
	log.Println("[fact_collaborations] extracting from PostgreSQL...")

	var rows []factCollaborationRow
	if err := db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return fmt.Errorf("fact_collaborations extract: %w", err)
	}
	log.Printf("[fact_collaborations] extracted %d rows", len(rows))

	if len(rows) == 0 {
		return nil
	}

	ctx := context.Background()
	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO fact_collaborations")
	if err != nil {
		return fmt.Errorf("fact_collaborations prepare batch: %w", err)
	}
	for _, r := range rows {
		dateID := toDateID(r.Date)
		if err := batch.Append(
			uint32(r.CollaborationID), r.Date, dateID,
			uint32(r.ApplicationID), uint32(r.EventID),
			uint32(r.CreatorID), uint32(r.VenueID),
			r.Status, r.CreatedAt, r.UpdatedAt,
		); err != nil {
			return fmt.Errorf("fact_collaborations append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("fact_collaborations send: %w", err)
	}

	log.Printf("[fact_collaborations] loaded %d rows into ClickHouse", len(rows))
	return nil
}
