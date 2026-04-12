package etl

import (
	"analytics-service/internal/clickhouse"
	"context"
	"fmt"
	"log"
	"time"
)

type dimDateRow struct {
	DateID      uint32
	Date        time.Time
	DayOfWeek   uint8
	DayOfMonth  uint8
	Month       uint8
	Quarter     uint8
	Year        uint16
	IsWeekend   uint8
}

// BuildDimDate генерирует календарь с 2020-01-01 по 2030-12-31.
// Пропускает если таблица уже заполнена.
func BuildDimDate(ch *clickhouse.Client) error {
	ctx := context.Background()

	var count uint64
	row := ch.Conn().QueryRow(ctx, "SELECT count() FROM dim_date")
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("dim_date count: %w", err)
	}
	if count > 0 {
		log.Println("[dim_date] already populated, skipping")
		return nil
	}

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)

	rows := make([]dimDateRow, 0, 4018)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dow := uint8(d.Weekday()) // 0=Sunday
		rows = append(rows, dimDateRow{
			DateID:     uint32(d.Year()*10000 + int(d.Month())*100 + d.Day()),
			Date:       d,
			DayOfWeek:  dow,
			DayOfMonth: uint8(d.Day()),
			Month:      uint8(d.Month()),
			Quarter:    uint8((int(d.Month())-1)/3 + 1),
			Year:       uint16(d.Year()),
			IsWeekend:  boolToUint8(dow == 0 || dow == 6),
		})
	}

	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO dim_date")
	if err != nil {
		return fmt.Errorf("dim_date prepare batch: %w", err)
	}
	for _, r := range rows {
		if err := batch.Append(r.DateID, r.Date, r.DayOfWeek, r.DayOfMonth,
			r.Month, r.Quarter, r.Year, r.IsWeekend); err != nil {
			return fmt.Errorf("dim_date append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("dim_date send: %w", err)
	}

	log.Printf("[dim_date] generated %d rows", len(rows))
	return nil
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
