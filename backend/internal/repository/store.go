package repository

import (
	"errors"
	"time"

	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListActiveCampaigns(category string, limit, offset int) ([]domain.Campaign, error) {
	query := s.db.Where("status = ?", "active").Order("created_at desc")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var campaigns []domain.Campaign
	err := query.Limit(limit).Offset(offset).Find(&campaigns).Error
	return campaigns, err
}

func (s *Store) GetActiveCampaignBySlug(slug string) (*domain.Campaign, error) {
	var campaign domain.Campaign
	err := s.db.Where("slug = ? AND status = ?", slug, "active").First(&campaign).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &campaign, err
}

func (s *Store) GetActiveCampaignByID(id string) (*domain.Campaign, error) {
	var campaign domain.Campaign
	err := s.db.Where("id = ? AND status = ?", id, "active").First(&campaign).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &campaign, err
}

func (s *Store) CreateCampaign(campaign *domain.Campaign) error {
	return s.db.Create(campaign).Error
}

func (s *Store) UpdateCampaign(campaign *domain.Campaign) error {
	return s.db.Model(&domain.Campaign{}).Where("id = ?", campaign.ID).Updates(map[string]any{
		"title":            campaign.Title,
		"slug":             campaign.Slug,
		"description":      campaign.Description,
		"category":         campaign.Category,
		"subcategory":      campaign.Subcategory,
		"campaign_type":    campaign.CampaignType,
		"banner_url":       campaign.BannerURL,
		"location":         campaign.Location,
		"beneficiary_note": campaign.BeneficiaryNote,
		"target_amount":    campaign.TargetAmount,
		"end_date":         campaign.EndDate,
	}).Error
}

func (s *Store) FindOrCreateDonor(donor *domain.Donor) (*domain.Donor, error) {
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "phone_number"}},
		DoNothing: true,
	}).Create(donor).Error; err != nil {
		return nil, err
	}

	var existing domain.Donor
	if err := s.db.Where("phone_number = ?", donor.PhoneNumber).First(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *Store) CreateDonation(donation *domain.Donation) error {
	return s.db.Create(donation).Error
}

func (s *Store) UpdateDonationProvider(donationID, providerOrderID, providerStatus string) error {
	return s.db.Model(&domain.Donation{}).Where("id = ?", donationID).Updates(map[string]any{
		"provider":          "midtrans",
		"provider_order_id": providerOrderID,
		"provider_status":   providerStatus,
	}).Error
}

func (s *Store) ApplyPaymentStatus(orderID, providerStatus, providerTransactionID, providerPayload, status string, paidAt *time.Time) (bool, error) {
	var counted bool
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var donation domain.Donation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_order_id = ?", orderID).First(&donation).Error; err != nil {
			return err
		}

		finalStatus := status
		if donation.Status == "success" {
			finalStatus = "success"
		}

		updates := map[string]any{
			"status":                  finalStatus,
			"provider_status":         providerStatus,
			"provider_transaction_id": providerTransactionID,
			"provider_payload":        providerPayload,
		}
		if status == "success" && paidAt != nil {
			updates["paid_at"] = paidAt
		}

		if status == "success" && donation.Status != "success" {
			if err := tx.Model(&domain.Campaign{}).Where("id = ?", donation.CampaignID).UpdateColumn("collected_amount", gorm.Expr("collected_amount + ?", donation.Amount)).Error; err != nil {
				return err
			}
			counted = true
		}

		return tx.Model(&domain.Donation{}).Where("id = ?", donation.ID).Updates(updates).Error
	})
	return counted, err
}

func (s *Store) GetDonation(id string) (*domain.Donation, error) {
	var donation domain.Donation
	err := s.db.Preload("Campaign").Preload("Donor").First(&donation, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &donation, err
}

func (s *Store) ListDonations(limit int) ([]domain.Donation, error) {
	var donations []domain.Donation
	err := s.db.Preload("Campaign").Preload("Donor").Order("created_at desc").Limit(limit).Find(&donations).Error
	return donations, err
}
