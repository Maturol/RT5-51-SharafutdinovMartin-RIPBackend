package ds

import "time"

type Bloodlosscalc struct {
	ID            int       `gorm:"primaryKey;column:bloodlosscalc_id"`
	Status        string    `gorm:"size:16;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	CreatorID     int       `gorm:"not null"`
	FormedAt      *time.Time
	CompletedAt   *time.Time
	ModeratorID   *int
	PatientHeight float64 `gorm:"type:numeric(5,2)"`
	PatientWeight int     `gorm:"not null"`
}
