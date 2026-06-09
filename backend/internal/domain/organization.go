package domain

import "time"

type Organization struct {
	ID           string     `gorm:"type:char(36);primaryKey;" json:"id"`
	Name         string     `gorm:"type:varchar(255);not null;" json:"name"`
	Email        string     `gorm:"type:varchar(191);uniqueIndex;not null;" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null;" json:"-"`
	Address      string     `gorm:"type:text;" json:"address"`
	Status       string     `gorm:"type:varchar(50);default:'pending';not null;" json:"status"` // pending, active, suspended
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index;" json:"-"`
	
	Campaigns    []Campaign `gorm:"foreignKey:OrganizationID" json:"campaigns,omitempty"`
}
