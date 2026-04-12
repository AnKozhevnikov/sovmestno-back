package clickhouse

import (
	"context"
	"fmt"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Client struct {
	conn driver.Conn
}

func NewClient(host, database, username, password string) (*Client, error) {
	conn, err := ch.Open(&ch.Options{
		Addr: []string{host},
		Auth: ch.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout: 10 * time.Second,
		Settings: ch.Settings{
			"max_execution_time": 60,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Conn() driver.Conn {
	return c.conn
}

func (c *Client) Exec(ctx context.Context, query string) error {
	return c.conn.Exec(ctx, query)
}

func (c *Client) InitSchemas() error {
	ctx := context.Background()

	schemas := []string{
		// --- Dimension tables ---
		`CREATE TABLE IF NOT EXISTS dim_date (
			date_id      UInt32,
			date         Date,
			day_of_week  UInt8,
			day_of_month UInt8,
			month        UInt8,
			quarter      UInt8,
			year         UInt16,
			is_weekend   UInt8
		) ENGINE = MergeTree()
		ORDER BY date`,

		`CREATE TABLE IF NOT EXISTS dim_user (
			user_id           UInt32,
			role              String,
			name              String,
			description       String,
			city              String,
			capacity          UInt32,
			registration_date Date,
			profile_filled    UInt8,
			updated_at        DateTime
		) ENGINE = ReplacingMergeTree(updated_at)
		ORDER BY user_id`,

		`CREATE TABLE IF NOT EXISTS dim_event (
			event_id     UInt32,
			creator_id   UInt32,
			title        String,
			category_ids Array(UInt32),
			is_active    UInt8,
			is_completed UInt8,
			created_date Date,
			updated_at   DateTime
		) ENGINE = ReplacingMergeTree(updated_at)
		ORDER BY event_id`,

		// --- Fact tables ---
		`CREATE TABLE IF NOT EXISTS fact_applications (
			application_id UInt32,
			date           Date,
			date_id        UInt32,
			sender_id      UInt32,
			sender_type    String,
			receiver_id    UInt32,
			receiver_type  String,
			event_id       UInt32,
			status         String,
			created_at     DateTime,
			updated_at     DateTime
		) ENGINE = ReplacingMergeTree(updated_at)
		ORDER BY (date, application_id)`,

		`CREATE TABLE IF NOT EXISTS fact_collaborations (
			collaboration_id UInt32,
			date             Date,
			date_id          UInt32,
			application_id   UInt32,
			event_id         UInt32,
			creator_id       UInt32,
			venue_id         UInt32,
			status           String,
			created_at       DateTime,
			updated_at       DateTime
		) ENGINE = ReplacingMergeTree(updated_at)
		ORDER BY (date, collaboration_id)`,
	}

	for _, ddl := range schemas {
		if err := c.conn.Exec(ctx, ddl); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}
	return nil
}
