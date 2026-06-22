package domain

import "time"

type SchemaMigration struct {
	Version   string    `gorm:"type:varchar(191);primaryKey;" json:"version"`
	AppliedAt time.Time `gorm:"not null;" json:"applied_at"`
}
