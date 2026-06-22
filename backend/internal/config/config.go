package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DBDriver           string
	DBDsn              string
	Env                string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFrom           string
	SMTPFromName       string
	SMTPEncryption     string
	WaitlistEmailURL   string
	WaitlistAdminEmail string
	CampaignAdminToken string
}

func Load() *Config {
	// Load .env file if it exists, ignore error if it doesn't
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbDsn := getEnv("DB_DSN", "kebaikanku.db")
	env := getEnv("APP_ENV", "development")

	return &Config{
		Port:               port,
		DBDriver:           dbDriver,
		DBDsn:              dbDsn,
		Env:                env,
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPass:           getEnv("SMTP_PASS", ""),
		SMTPFrom:           getEnv("SMTP_FROM", ""),
		SMTPFromName:       getEnv("SMTP_FROM_NAME", "kebaikanku.id"),
		SMTPEncryption:     getEnv("SMTP_ENCRYPTION", ""),
		WaitlistEmailURL:   getEnv("WAITLIST_EMAIL_URL", "https://kebaikanku.id/coming-soon"),
		WaitlistAdminEmail: getEnv("WAITLIST_ADMIN_EMAIL", ""),
		CampaignAdminToken: getEnv("CAMPAIGN_ADMIN_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
