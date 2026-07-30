package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) *gorm.DB {
	var err error
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch cfg.DBDriver {
	case "postgres":
		DB, err = gorm.Open(postgres.Open(cfg.DBDsn), gormConfig)
	case "sqlite", "":
		DB, err = gorm.Open(sqlite.Open(cfg.DBDsn), gormConfig)
	default:
		log.Fatalf("Unsupported database driver: %s", cfg.DBDriver)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Printf("Database connected successfully using driver: %s\n", cfg.DBDriver)

	if cfg.Env == "production" {
		fmt.Println("Skipping GORM AutoMigrate in production; run tracked SQL migrations before starting the API.")
		return DB
	}

	// AutoMigrate keeps SQLite-based local development friction-free. Production
	// schema changes are applied by the tracked SQL migrations in backend/migrations.
	err = DB.AutoMigrate(
		&domain.SchemaMigration{},
		&domain.Organization{},
		&domain.Campaign{},
		&domain.Donor{},
		&domain.Donation{},
		&domain.Waitlist{},
		&domain.PaymentSetting{},
	)
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	fmt.Println("Database schema migrated successfully.")

	recordSchemaVersion(DB, "20260622_alpha_mvp")

	return DB
}

const defaultOrganizationID = "00000000-0000-4000-8000-000000000001"

// SeedDefaults creates only the records required for a useful first boot.
func SeedDefaults(db *gorm.DB, adminPassword string) (string, error) {
	var count int64
	if err := db.Model(&domain.Organization{}).Count(&count).Error; err != nil || count > 0 {
		return "", err
	}
	initialPassword := adminPassword
	generated := false
	if initialPassword == "" {
		secret := make([]byte, 18)
		if _, err := rand.Read(secret); err != nil {
			return "", err
		}
		initialPassword = base64.RawURLEncoding.EncodeToString(secret)
		generated = true
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	endDate := time.Now().UTC().AddDate(1, 0, 0)
	err = db.Transaction(func(tx *gorm.DB) error {
		organization := domain.Organization{ID: defaultOrganizationID, Name: "kebaikanku.id", Email: "admin@kebaikanku.id", PasswordHash: string(hash), Status: "active"}
		if err := tx.Create(&organization).Error; err != nil {
			return err
		}
		return tx.Create(&domain.Campaign{
			ID: "00000000-0000-4000-8000-000000000002", OrganizationID: organization.ID,
			Title: "Bantu Sesama Hari Ini", Slug: "bantu-sesama-hari-ini", Category: "kemanusiaan",
			Description: "Campaign contoh siap diedit dari dashboard admin.", TargetAmount: 10_000_000,
			EndDate: endDate, Status: "active", CampaignType: "target_deadline",
		}).Error
	})
	if err != nil {
		return "", err
	}
	if generated {
		return initialPassword, nil
	}
	return "", nil
}

func recordSchemaVersion(db *gorm.DB, version string) {
	migration := domain.SchemaMigration{
		Version:   version,
		AppliedAt: time.Now().UTC(),
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "version"}},
		DoNothing: true,
	}).Create(&migration).Error; err != nil {
		log.Fatalf("Failed to record schema migration %s: %v", version, err)
	}
}
