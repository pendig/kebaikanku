package database

import (
	"testing"

	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedDefaultsIsSafeAndGeneratesAdminPassword(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:seed-defaults?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Organization{}, &domain.Campaign{}); err != nil {
		t.Fatal(err)
	}
	password, err := SeedDefaults(db, "")
	if err != nil || len(password) < 20 {
		t.Fatalf("password length=%d err=%v", len(password), err)
	}
	var organization domain.Organization
	if err := db.First(&organization).Error; err != nil {
		t.Fatal(err)
	}
	if organization.PasswordHash == password || bcrypt.CompareHashAndPassword([]byte(organization.PasswordHash), []byte(password)) != nil {
		t.Fatal("generated password was not stored as a bcrypt hash")
	}
	if second, err := SeedDefaults(db, ""); err != nil || second != "" {
		t.Fatalf("second seed=%q err=%v", second, err)
	}
	var campaigns int64
	if err := db.Model(&domain.Campaign{}).Count(&campaigns).Error; err != nil || campaigns != 1 {
		t.Fatalf("campaigns=%d err=%v", campaigns, err)
	}
}
