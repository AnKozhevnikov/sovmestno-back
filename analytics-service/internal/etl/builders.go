package etl

import (
	"analytics-service/internal/clickhouse"
	"log"
	"time"

	"gorm.io/gorm"
)

type Builders struct {
	db *gorm.DB
	ch *clickhouse.Client
}

func NewBuilders(db *gorm.DB, ch *clickhouse.Client) *Builders {
	return &Builders{db: db, ch: ch}
}

// RunFull полная перестройка всех таблиц. Запускается при старте сервиса.
func (b *Builders) RunFull() {
	run("dim_date", func() error { return BuildDimDate(b.ch) })
	run("dim_user", func() error { return BuildDimUser(b.db, b.ch) })
	run("dim_event", func() error { return BuildDimEvent(b.db, b.ch) })
	run("fact_applications", func() error { return BuildFactApplicationsAll(b.db, b.ch) })
	run("fact_collaborations", func() error { return BuildFactCollaborationsAll(b.db, b.ch) })
}

func (b *Builders) RunDaily() {
	since := time.Now().AddDate(0, 0, -2)
	run("dim_user", func() error { return BuildDimUser(b.db, b.ch) })
	run("dim_event", func() error { return BuildDimEvent(b.db, b.ch) })
	run("fact_applications", func() error { return BuildFactApplicationsIncremental(b.db, b.ch, since) })
	run("fact_collaborations", func() error { return BuildFactCollaborationsIncremental(b.db, b.ch, since) })
}

func run(name string, fn func() error) {
	log.Printf("[etl] starting %s", name)
	if err := fn(); err != nil {
		log.Printf("[etl] %s failed: %v", name, err)
	} else {
		log.Printf("[etl] %s done", name)
	}
}
