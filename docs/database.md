# Database Architecture & GORM Integration

This document outlines the database design, model structures, and how **kebaikanku.id** implements dynamic database switching (SQLite for local/SaaS trials, PostgreSQL for enterprise deployments).

---

## 🔌 Dynamic Database Driver Loading

We use **GORM** (Go Object Relational Mapper) because it abstracts dialect-specific SQL logic, allowing us to swap the underlying database driver simply by changing environment variables.

### Environment Variables
```env
# Database Driver Selection: "sqlite" or "postgres"
DB_DRIVER=sqlite

# DSN for SQLite
DB_DSN=kebaikanku.db

# DSN for PostgreSQL (Example)
# DB_DRIVER=postgres
# DB_DSN=host=localhost user=gorm password=gorm dbname=kebaikanku port=5432 sslmode=disable TimeZone=Asia/Jakarta
```

### Connection Initialization (Go Implementation Concept)
```go
package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")

	switch driver {
	case "postgres":
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite", "":
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		log.Fatalf("Unsupported database driver: %s", driver)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Printf("Database connected successfully using driver: %s\n", driver)
}
```

---

## 📐 Schema Models (Entities)

To ensure smooth operation on both SQLite and PostgreSQL, we use compatible data types. For example, SQLite does not have a native UUID type, so UUIDs are stored as text (char(36)), which works perfectly in both systems.

### 1. Organization (Lembaga Zakat / Yayasan)
Stores details of the registered non-profit institution.
```go
type Organization struct {
	ID           string    `gorm:"type:char(36);primaryKey;" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null;" json:"name"`
	Email        string    `gorm:"type:varchar(191);uniqueIndex;not null;" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null;" json:"-"`
	Address      string    `gorm:"type:text;" json:"address"`
	Status       string    `gorm:"type:varchar(50);default:'pending';not null;" json:"status"` // pending, active, suspended
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	Campaigns    []Campaign `gorm:"foreignKey:OrganizationID" json:"campaigns,omitempty"`
}
```

### 2. Campaign (Crowdfunding Projects)
Represents fundraising initiatives.
```go
type Campaign struct {
	ID             string    `gorm:"type:char(36);primaryKey;" json:"id"`
	OrganizationID string    `gorm:"type:char(36);not null;index;" json:"organization_id"`
	Title          string    `gorm:"type:varchar(255);not null;" json:"title"`
	Slug           string    `gorm:"type:varchar(255);uniqueIndex;not null;" json:"slug"`
	Description    string    `gorm:"type:text;" json:"description"`
	Category       string    `gorm:"type:varchar(100);not null;" json:"category"` // zakat_maal, zakat_fitrah, infak, kemanusiaan
	TargetAmount   float64   `gorm:"type:decimal(15,2);not null;" json:"target_amount"`
	CollectedAmount float64  `gorm:"type:decimal(15,2);default:0.00;not null;" json:"collected_amount"`
	EndDate        time.Time `json:"end_date"`
	Status         string    `gorm:"type:varchar(50);default:'active';not null;" json:"status"` // active, completed, paused
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
```

### 3. Donor
Stores donor contact profiles.
```go
type Donor struct {
	ID          string `gorm:"type:char(36);primaryKey;" json:"id"`
	Name        string `gorm:"type:varchar(255);not null;" json:"name"`
	PhoneNumber string `gorm:"type:varchar(50);uniqueIndex;not null;" json:"phone_number"`
	Email       string `gorm:"type:varchar(191);" json:"email"`
}
```

### 4. Donation
Represents a payment transaction.
```go
type Donation struct {
	ID            string    `gorm:"type:char(36);primaryKey;" json:"id"`
	CampaignID    string    `gorm:"type:char(36);not null;index;" json:"campaign_id"`
	DonorID       string    `gorm:"type:char(36);not null;index;" json:"donor_id"`
	Amount        float64   `gorm:"type:decimal(15,2);not null;" json:"amount"`
	PlatformTip   float64   `gorm:"type:decimal(15,2);default:0.00;" json:"platform_tip"`
	PGFee         float64   `gorm:"type:decimal(15,2);default:0.00;" json:"pg_fee"`
	Status        string    `gorm:"type:varchar(50);default:'pending';not null;" json:"status"` // pending, success, failed
	PaymentMethod string    `gorm:"type:varchar(100);" json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
	
	Campaign      Campaign  `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	Donor         Donor     `gorm:"foreignKey:DonorID" json:"donor,omitempty"`
}
```

---

## 🔄 Schema Migrations

Local development uses GORM's `AutoMigrate` command at startup. Production must use the ordered SQL migrations in [`backend/migrations`](../backend/migrations), which are applied by the production Compose migration job and recorded in `kebaikanku_migrations`.

```go
func RunMigrations() {
	err := DB.AutoMigrate(
		&Organization{},
		&Campaign{},
		&Donor{},
		&Donation{},
	)
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	fmt.Println("Database schema migrated successfully.")
}
```

> [!WARNING]
> `APP_ENV=production` disables AutoMigrate. Create a new forward-only migration for every production schema change and take a backup before applying it. Do not alter migration files that have already been deployed.

Current tracked schema version:

```text
000001_initial_schema
000002_campaign_metadata
000003_donation_checkout_idempotency
```

This is the initial production baseline. A production rollback restores the last verified database backup and deploys the prior API image; donations make destructive down migrations unsafe.
