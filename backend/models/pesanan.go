package models

import "time"

type Pesanan struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	Nama      string        `json:"nama" gorm:"size:100;not null"`
	NoMeja    int           `json:"no_meja" gorm:"not null"`
	Status    string        `json:"status" gorm:"size:20;default:diproses"`
	Total     int           `json:"total"`
	CreatedAt time.Time     `json:"tanggal"`
	Items     []PesananItem `json:"items" gorm:"foreignKey:PesananID"`
}

type PesananItem struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PesananID  uint   `json:"pesanan_id" gorm:"not null;index"`
	ProductID  uint   `json:"product_id"`
	Nama       string `json:"nama"`
	Harga      int    `json:"harga"`
	Jumlah     int    `json:"jumlah"`
	Keterangan string `json:"keterangan"`
}