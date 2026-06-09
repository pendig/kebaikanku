package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	DBDriver string
	DBDsn    string
	Env      string
}

func Load() *Config {
	// Load .env file if it exists, ignore error if it doesn't
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbDsn := getEnv("DB_DSN", "kebaikanku.db")
	env := getEnv("APP_ENV", "development")

	return &Config{
		Port:     port,
		DBDriver: dbDriver,
		DBDsn:    dbDsn,
		Env:      env,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
