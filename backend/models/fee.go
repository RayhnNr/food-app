package models

import "time"

type Fee struct {
	ID        uint      `gorm:"column:fee_id;primaryKey;autoIncrement" json:"fee_id"`
	Slug      string    `gorm:"column:fee_slug;size:256;not null;uniqueIndex:idx_fee_user_slug,priority:2" json:"fee_slug"`
	Name      string    `gorm:"column:fee_name;size:256;not null" json:"fee_name"`
	Cost      float64   `gorm:"column:fee_cost;not null;default:0" json:"fee_cost"`
	UserID    uint      `gorm:"column:user_id;not null;uniqueIndex:idx_fee_user_slug,priority:1" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (Fee) TableName() string {
	return "fee"
}