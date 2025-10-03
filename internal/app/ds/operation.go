package ds

type Operation struct {
	ID             int     `gorm:"primaryKey;column:operation_id"`
	Title          string  `gorm:"size:64;not null"`
	Description    string  `gorm:"type:text;not null"`
	Status         string  `gorm:"size:16;not null"`
	ImageURL       *string `gorm:"size:256"`
	BloodLossCoeff float64 `gorm:"type:numeric(5,2);not null"`
	AvgBloodLoss   int     `gorm:"not null"`
}
