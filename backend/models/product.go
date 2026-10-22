package models

type Product struct {
	ID uint `json:"id" gorm:"primaryKey"`
	Kode  string `json:"kode" gorm:"uniqueIndex;size:20;not null"`
	Nama string `json:"nama" gorm:"size:100;not null"`
	CategoryID *uint     `json:"category_id"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Harga int `json:"harga" gorm:"not null"`
	IsReady bool `json:"is_ready" gorm:"default:true"`
	Gambar string `json:"gambar"`
}