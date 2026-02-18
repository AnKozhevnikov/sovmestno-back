package config

import "os"

type Config struct {
	Port           string
	GinMode        string
	DatabaseDSN    string
	JWTSecret      string
	AdminSecretKey string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8081"),
		GinMode:        getEnv("GIN_MODE", "release"),
		DatabaseDSN:    getEnv("DB_DSN", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AdminSecretKey: getEnv("ADMIN_SECRET_KEY", ""),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", ""),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
