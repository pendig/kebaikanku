package domain

import "time"

type Campaign struct {
	ID              string     `gorm:"type:char(36);primaryKey;" json:"id"`
	OrganizationID  string     `gorm:"type:char(36);not null;index;" json:"organization_id"`
	Title           string     `gorm:"type:varchar(255);not null;" json:"title"`
	Slug            string     `gorm:"type:varchar(255);uniqueIndex;not null;" json:"slug"`
	Description     string     `gorm:"type:text;" json:"description"`
	Category        string     `gorm:"type:varchar(100);not null;" json:"category"` // zakat_maal, zakat_fitrah, infak, kemanusiaan
	TargetAmount    float64    `gorm:"type:decimal(15,2);not null;" json:"target_amount"`
	CollectedAmount float64    `gorm:"type:decimal(15,2);default:0.00;not null;" json:"collected_amount"`
	EndDate         time.Time  `json:"end_date"`
	Status          string     `gorm:"type:varchar(50);default:'active';not null;" json:"status"` // active, completed, paused
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index;" json:"-"`
}
