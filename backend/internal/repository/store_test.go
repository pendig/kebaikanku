package repository

import (
	"errors"
	"testing"

	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateCampaignReturnsNotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Campaign{}); err != nil {
		t.Fatal(err)
	}

	err = NewStore(db).UpdateCampaign(&domain.Campaign{ID: "missing"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("UpdateCampaign() error = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestApplyPaymentStatusCountsSuccessOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Organization{}, &domain.Campaign{}, &domain.Donor{}, &domain.Donation{}); err != nil {
		t.Fatal(err)
	}

	campaign := domain.Campaign{ID: "campaign-1", OrganizationID: "org-1", Title: "Campaign", Slug: "campaign", Category: "infak", TargetAmount: 1000, Status: "active"}
	donor := domain.Donor{ID: "donor-1", Name: "Budi", PhoneNumber: "+6281"}
	donation := domain.Donation{ID: "donation-1", CampaignID: campaign.ID, DonorID: donor.ID, Amount: 100, Status: "pending", ProviderOrderID: "order-1"}
	if err := db.Create(&campaign).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&donor).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&donation).Error; err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)
	counted, err := store.ApplyPaymentStatus("order-1", "settlement", "trx-1", "{}", "success", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !counted {
		t.Fatal("first success should be counted")
	}
	counted, err = store.ApplyPaymentStatus("order-1", "settlement", "trx-1", "{}", "success", nil)
	if err != nil {
		t.Fatal(err)
	}
	if counted {
		t.Fatal("duplicate success should not be counted")
	}
	counted, err = store.ApplyPaymentStatus("order-1", "expire", "trx-1", "{}", "failed", nil)
	if err != nil {
		t.Fatal(err)
	}
	if counted {
		t.Fatal("failed callback after success should not be counted")
	}

	var updated domain.Campaign
	if err := db.First(&updated, "id = ?", campaign.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.CollectedAmount != 100 {
		t.Fatalf("collected amount = %v, want 100", updated.CollectedAmount)
	}
	var updatedDonation domain.Donation
	if err := db.First(&updatedDonation, "id = ?", donation.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updatedDonation.Status != "success" {
		t.Fatalf("donation status = %s, want success", updatedDonation.Status)
	}
}
