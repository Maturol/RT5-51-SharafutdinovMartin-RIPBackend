package ds

type BloodlosscalcOperation struct {
	BloodlosscalcID int `gorm:"primaryKey;column:bloodlosscalc_id"`
	OperationID     int `gorm:"primaryKey;column:operation_id"`
	HbBefore        *int
	HbAfter         *int
	SurgeryDuration *float64 `gorm:"type:numeric(4,1)"`
	TotalBloodLoss  *int
}
