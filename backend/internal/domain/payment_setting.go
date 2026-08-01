package domain

import "time"

type PaymentSetting struct {
	ID              string `gorm:"type:varchar(32);primaryKey"`
	Mode            string `gorm:"type:varchar(16);not null"`
	ServerKeyCipher string `gorm:"type:text"`
	ClientKeyCipher string `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
