package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "sqlite3://policy.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
