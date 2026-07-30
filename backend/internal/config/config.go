package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                       string
	DBDriver                   string
	DBDsn                      string
	Env                        string
	SMTPHost                   string
	SMTPPort                   string
	SMTPUser                   string
	SMTPPass                   string
	SMTPFrom                   string
	SMTPFromName               string
	SMTPEncryption             string
	WaitlistEmailURL           string
	WaitlistAdminEmail         string
	AdminPassword              string
	AdminSessionSecret         string
	AdminSettingsEncryptionKey string
	UploadDir                  string
	PublicUploadBaseURL        string
	PublicLandingURL           string
	MidtransEnv                string
	MidtransServerKey          string
	MidtransClientKey          string
	MidtransNotifyKey          string
	CORSAllowedOrigins         string
	PublicRateLimit            int
}

func Load() *Config {
	// Load .env file if it exists, ignore error if it doesn't
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbDsn := getEnv("DB_DSN", "kebaikanku.db")
	env := getEnv("APP_ENV", "development")

	return &Config{
		Port:                       port,
		DBDriver:                   dbDriver,
		DBDsn:                      dbDsn,
		Env:                        env,
		SMTPHost:                   getEnv("SMTP_HOST", ""),
		SMTPPort:                   getEnv("SMTP_PORT", "587"),
		SMTPUser:                   getEnv("SMTP_USER", ""),
		SMTPPass:                   getEnv("SMTP_PASS", ""),
		SMTPFrom:                   getEnv("SMTP_FROM", ""),
		SMTPFromName:               getEnv("SMTP_FROM_NAME", "kebaikanku.id"),
		SMTPEncryption:             getEnv("SMTP_ENCRYPTION", ""),
		WaitlistEmailURL:           getEnv("WAITLIST_EMAIL_URL", "https://kebaikanku.id/coming-soon"),
		WaitlistAdminEmail:         getEnv("WAITLIST_ADMIN_EMAIL", ""),
		AdminPassword:              getEnv("ADMIN_PASSWORD", ""),
		AdminSessionSecret:         getEnv("ADMIN_SESSION_SECRET", ""),
		AdminSettingsEncryptionKey: getEnv("ADMIN_SETTINGS_ENCRYPTION_KEY", ""),
		UploadDir:                  getEnv("UPLOAD_DIR", "uploads"),
		PublicUploadBaseURL:        getEnv("PUBLIC_UPLOAD_BASE_URL", "/uploads"),
		PublicLandingURL:           getEnv("PUBLIC_LANDING_URL", "http://127.0.0.1:18481"),
		MidtransEnv:                getEnv("MIDTRANS_ENV", "sandbox"),
		MidtransServerKey:          getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:          getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransNotifyKey:          getEnv("MIDTRANS_NOTIFICATION_TOKEN", ""),
		CORSAllowedOrigins:         getEnv("CORS_ALLOWED_ORIGINS", ""),
		PublicRateLimit:            getEnvInt("PUBLIC_RATE_LIMIT_PER_MINUTE", 120),
	}
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
