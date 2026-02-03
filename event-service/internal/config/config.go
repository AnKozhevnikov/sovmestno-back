package config

import "os"

type Config struct {
	Port        string
	GinMode     string
	DatabaseDSN string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8082"),
		GinMode:     getEnv("GIN_MODE", "release"),
		DatabaseDSN: getEnv("DB_DSN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
