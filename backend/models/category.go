package models

import "time"

type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"size:50;uniqueIndex;not null"`
	Deskripsi string    `json:"deskripsi"`
	CreatedAt time.Time `json:"created_at"`
}

