package main

import (
	"log"
	"os"
	"strings"

	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/database"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()
	database.Init(cfg)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(requiredSecret("ADMIN_PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash admin password: %v", err)
	}
	organization := domain.Organization{
		ID:           value("PILOT_ORGANIZATION_ID", "pilot-org"),
		Name:         required("PILOT_ORGANIZATION_NAME"),
		Email:        required("PILOT_ORGANIZATION_EMAIL"),
		PasswordHash: string(passwordHash),
		Status:       "active",
	}
	if err := database.DB.Where("id = ?", organization.ID).FirstOrCreate(&organization).Error; err != nil {
		log.Fatalf("bootstrap organization: %v", err)
	}
	log.Printf("pilot organization %q is ready", organization.ID)
}

func value(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func required(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func requiredSecret(key string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}
