package config

import "os"

type Config struct {
	Port             string
	PostgresDSN      string
	ClickHouseHost   string
	ClickHouseDB     string
	ClickHouseUser   string
	ClickHousePass   string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8084"),
		PostgresDSN:    getEnv("DB_DSN", ""),
		ClickHouseHost: getEnv("CLICKHOUSE_HOST", "clickhouse:9000"),
		ClickHouseDB:   getEnv("CLICKHOUSE_DB", "analytics_db"),
		ClickHouseUser: getEnv("CLICKHOUSE_USER", "analytics_user"),
		ClickHousePass: getEnv("CLICKHOUSE_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
