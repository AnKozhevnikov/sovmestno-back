package etl

import (
	"analytics-service/internal/clickhouse"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type factApplicationRow struct {
	ApplicationID int       `gorm:"column:application_id"`
	Date          time.Time `gorm:"column:date"`
	SenderID      int       `gorm:"column:sender_id"`
	SenderType    string    `gorm:"column:sender_type"`
	ReceiverID    int       `gorm:"column:receiver_id"`
	ReceiverType  string    `gorm:"column:receiver_type"`
	EventID       int       `gorm:"column:event_id"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

const factApplicationsQuery = `
SELECT id AS application_id, DATE(created_at) AS date,
       sender_id, sender_type, receiver_id, receiver_type,
       event_id, status, created_at, updated_at
FROM applications
`

const factApplicationsIncrementalQuery = `
SELECT id AS application_id, DATE(created_at) AS date,
       sender_id, sender_type, receiver_id, receiver_type,
       event_id, status, created_at, updated_at
FROM applications
WHERE updated_at >= ?
`

// BuildFactApplicationsAll полная перестройка.
func BuildFactApplicationsAll(db *gorm.DB, ch *clickhouse.Client) error {
	return buildFactApplications(db, ch, factApplicationsQuery)
}

// BuildFactApplicationsIncremental обновляет строки, изменившиеся с указанной даты.
func BuildFactApplicationsIncremental(db *gorm.DB, ch *clickhouse.Client, since time.Time) error {
	return buildFactApplications(db, ch, factApplicationsIncrementalQuery, since.Format("2006-01-02"))
}

func buildFactApplications(db *gorm.DB, ch *clickhouse.Client, query string, args ...interface{}) error {
	log.Println("[fact_applications] extracting from PostgreSQL...")

	var rows []factApplicationRow
	if err := db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return fmt.Errorf("fact_applications extract: %w", err)
	}
	log.Printf("[fact_applications] extracted %d rows", len(rows))

	if len(rows) == 0 {
		return nil
	}

	ctx := context.Background()
	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO fact_applications")
	if err != nil {
		return fmt.Errorf("fact_applications prepare batch: %w", err)
	}
	for _, r := range rows {
		dateID := toDateID(r.Date)
		if err := batch.Append(
			uint32(r.ApplicationID), r.Date, dateID,
			uint32(r.SenderID), r.SenderType,
			uint32(r.ReceiverID), r.ReceiverType,
			uint32(r.EventID), r.Status,
			r.CreatedAt, r.UpdatedAt,
		); err != nil {
			return fmt.Errorf("fact_applications append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("fact_applications send: %w", err)
	}

	log.Printf("[fact_applications] loaded %d rows into ClickHouse", len(rows))
	return nil
}

func toDateID(t time.Time) uint32 {
	return uint32(t.Year()*10000 + int(t.Month())*100 + t.Day())
}
