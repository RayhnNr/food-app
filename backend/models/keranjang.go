package models

type Keranjang struct{
	ID         uint    `json:"id" gorm:"primaryKey"`
	ProductID  uint    `json:"product_id" gorm:"not null"`
	Product    Product `json:"product" gorm:"foreignKey:ProductID"`
	Jumlah     int     `json:"jumlah" gorm:"not null;default:1"`
	Keterangan string  `json:"keterangan"`
}