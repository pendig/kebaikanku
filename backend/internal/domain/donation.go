package domain

import "time"

type Donation struct {
	ID                    string     `gorm:"type:char(36);primaryKey;" json:"id"`
	CampaignID            string     `gorm:"type:char(36);not null;index;" json:"campaign_id"`
	DonorID               string     `gorm:"type:char(36);not null;index;" json:"donor_id"`
	Amount                float64    `gorm:"type:decimal(15,2);not null;" json:"amount"`
	PlatformTip           float64    `gorm:"type:decimal(15,2);default:0.00;" json:"platform_tip"`
	PGFee                 float64    `gorm:"type:decimal(15,2);default:0.00;" json:"pg_fee"`
	Status                string     `gorm:"type:varchar(50);default:'pending';not null;" json:"status"` // pending, success, failed
	PaymentMethod         string     `gorm:"type:varchar(100);" json:"payment_method"`
	Provider              string     `gorm:"type:varchar(100);" json:"provider"`
	ProviderOrderID       string     `gorm:"type:varchar(191);index;" json:"provider_order_id"`
	IdempotencyKey        *string    `gorm:"type:varchar(255);uniqueIndex;" json:"-"`
	CheckoutToken         string     `gorm:"type:text;" json:"-"`
	CheckoutRedirectURL   string     `gorm:"type:text;" json:"-"`
	ProviderTransactionID string     `gorm:"type:varchar(191);index;" json:"provider_transaction_id"`
	ProviderStatus        string     `gorm:"type:varchar(100);" json:"provider_status"`
	ProviderPayload       string     `gorm:"type:text;" json:"-"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	Campaign *Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	Donor    *Donor    `gorm:"foreignKey:DonorID" json:"donor,omitempty"`
}
