package etl

import (
	"analytics-service/internal/clickhouse"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type dimUserRow struct {
	UserID           int       `gorm:"column:user_id"`
	Role             string    `gorm:"column:role"`
	Name             string    `gorm:"column:name"`
	Description      string    `gorm:"column:description"`
	City             string    `gorm:"column:city"`
	Capacity         int       `gorm:"column:capacity"`
	RegistrationDate time.Time `gorm:"column:registration_date"`
	ProfileFilled    int       `gorm:"column:profile_filled"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

const dimUserQuery = `
SELECT
    u.id                                                              AS user_id,
    u.role,
    COALESCE(NULLIF(c.name, ''), NULLIF(v.name, ''), '')             AS name,
    COALESCE(NULLIF(c.description, ''), NULLIF(v.description, ''), '') AS description,
    COALESCE(ci.name, '')                                             AS city,
    COALESCE(v.capacity, 0)                                           AS capacity,
    DATE(u.created_at)                                                AS registration_date,
    CASE
        WHEN COALESCE(NULLIF(c.name, ''), NULLIF(v.name, ''))        IS NOT NULL
         AND COALESCE(NULLIF(c.description, ''), NULLIF(v.description, '')) IS NOT NULL
        THEN 1 ELSE 0
    END                                                               AS profile_filled,
    GREATEST(
        u.updated_at,
        COALESCE(c.updated_at, v.updated_at, u.updated_at)
    )                                                                 AS updated_at
FROM users u
LEFT JOIN creators c  ON c.user_id  = u.id
LEFT JOIN venues   v  ON v.user_id  = u.id
LEFT JOIN cities   ci ON ci.id      = v.city_id
WHERE u.role IN ('creator', 'venue')
`

func BuildDimUser(db *gorm.DB, ch *clickhouse.Client) error {
	log.Println("[dim_user] extracting from PostgreSQL...")

	var rows []dimUserRow
	if err := db.Raw(dimUserQuery).Scan(&rows).Error; err != nil {
		return fmt.Errorf("dim_user extract: %w", err)
	}
	log.Printf("[dim_user] extracted %d rows", len(rows))

	ctx := context.Background()

	if err := ch.Exec(ctx, "TRUNCATE TABLE dim_user"); err != nil {
		return fmt.Errorf("dim_user truncate: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO dim_user")
	if err != nil {
		return fmt.Errorf("dim_user prepare batch: %w", err)
	}
	for _, r := range rows {
		if err := batch.Append(
			uint32(r.UserID), r.Role, r.Name, r.Description,
			r.City, uint32(r.Capacity), r.RegistrationDate,
			uint8(r.ProfileFilled), r.UpdatedAt,
		); err != nil {
			return fmt.Errorf("dim_user append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("dim_user send: %w", err)
	}

	log.Printf("[dim_user] loaded %d rows into ClickHouse", len(rows))
	return nil
}
