package domain

import "time"

type Donor struct {
	ID          string     `gorm:"type:char(36);primaryKey;" json:"id"`
	Name        string     `gorm:"type:varchar(255);not null;" json:"name"`
	PhoneNumber string     `gorm:"type:varchar(50);uniqueIndex;not null;" json:"phone_number"`
	Email       string     `gorm:"type:varchar(191);" json:"email"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index;" json:"-"`
}
