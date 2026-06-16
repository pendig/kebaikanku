package domain

import "time"

type Waitlist struct {
	ID        string `gorm:"type:char(36);primaryKey;" json:"id"`
	Email     string `gorm:"type:varchar(191);uniqueIndex;not null;" json:"email"`
	Source    string `gorm:"type:varchar(255);" json:"source"`
	IPAddress string `gorm:"type:varchar(64);" json:"ip_address"`
	UserAgent string `gorm:"type:text;" json:"user_agent"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

