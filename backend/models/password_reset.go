package models

import "time"

type PasswordResetOTP struct {
	ID        uint       `json:"-" gorm:"primaryKey"`
	UserID    uint       `json:"-" gorm:"uniqueIndex;not null"`
	OTPHash   string     `json:"-" gorm:"not null"`
	ExpiresAt time.Time  `json:"-" gorm:"not null;index"`
	Attempts  int        `json:"-" gorm:"not null;default:0"`
	CreatedAt time.Time  `json:"-" gorm:"not null"`
	UsedAt    *time.Time `json:"-" gorm:"index"`
}

func (PasswordResetOTP) TableName() string {
	return "password_reset_otps"
}
