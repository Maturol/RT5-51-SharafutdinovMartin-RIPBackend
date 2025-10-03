package ds

type User struct {
	UserID       int    `gorm:"primaryKey;column:user_id"`
	Username     string `gorm:"size:32;unique;not null"`
	PasswordHash string `gorm:"size:128;not null"`
	IsModerator  bool   `gorm:"not null;default:false"`
}
