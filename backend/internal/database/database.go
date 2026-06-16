package database

import (
	"fmt"
	"log"

	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	// Run auto migrations
	err = DB.AutoMigrate(
		&domain.Organization{},
		&domain.Campaign{},
		&domain.Donor{},
		&domain.Donation{},
		&domain.Waitlist{},
	)
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	fmt.Println("Database schema migrated successfully.")

	return DB
}
