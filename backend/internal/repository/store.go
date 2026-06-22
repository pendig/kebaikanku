package repository

import (
	"errors"

	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"gorm.io/gorm"
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

func (s *Store) FindOrCreateDonor(donor *domain.Donor) (*domain.Donor, error) {
	var existing domain.Donor
	err := s.db.Where("phone_number = ?", donor.PhoneNumber).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return donor, s.db.Create(donor).Error
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

func (s *Store) GetDonation(id string) (*domain.Donation, error) {
	var donation domain.Donation
	err := s.db.Preload("Campaign").Preload("Donor").First(&donation, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &donation, err
}
