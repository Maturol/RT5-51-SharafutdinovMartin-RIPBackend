package ds

type BloodlosscalcOperation struct {
	BloodlosscalcID int     `gorm:"primaryKey;column:bloodlosscalc_id"`
	OperationID     int     `gorm:"primaryKey;column:operation_id"`
	HbBefore        int     `gorm:"not null"`
	HbAfter         int     `gorm:"not null"`
	SurgeryDuration float64 `gorm:"type:numeric(4,1);not null"`
	TotalBloodLoss  int     `gorm:"not null"`
}
