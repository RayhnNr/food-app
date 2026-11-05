package models

import "time"

// Bandul = data bandul (perantara/pemasok). Wilayah memakai tabel province, city, dan subdistrict.
// IDProvince, IDCity, IDSubdistrict kosong (NULL) = belum diisi.
//
// Nama field foreign key (IDProvince, IDCity, IDSubdistrict) sengaja berbeda dari primary key
// tujuan (ID), supaya GORM tidak salah membaca relasi.
type Bandul struct {
	ID            uint       `gorm:"column:bandul_id;primaryKey;autoIncrement" json:"bandul_id"`
	DateJoin      time.Time  `gorm:"column:bandul_date_join;not null" json:"bandul_date_join"`
	DateBirth     *time.Time `gorm:"column:bandul_date_birth;type:date" json:"bandul_date_birth"`
	Name          string     `gorm:"column:bandul_name;size:256;not null;index" json:"bandul_name"`
	Desc          string     `gorm:"column:bandul_desc;type:text" json:"bandul_desc"`
	Phone         string     `gorm:"column:bandul_phone;size:50" json:"bandul_phone"`
	MobilePhone   string     `gorm:"column:bandul_mobile_phone;size:50" json:"bandul_mobile_phone"`
	IDProvince    *uint      `gorm:"column:province_id;index" json:"province_id"`
	IDCity        *uint      `gorm:"column:city_id;index" json:"city_id"`
	IDSubdistrict *uint      `gorm:"column:subdistrict_id;index" json:"subdistrict_id"`
	Address       string     `gorm:"column:bandul_address;type:text" json:"bandul_address"`
	Active        int        `gorm:"column:bandul_active;not null;default:1" json:"bandul_active"`
	UserID        uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	User        *User        `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Province    *Province    `gorm:"foreignKey:IDProvince;references:ID" json:"-"`
	City        *City        `gorm:"foreignKey:IDCity;references:ID" json:"-"`
	Subdistrict *Subdistrict `gorm:"foreignKey:IDSubdistrict;references:ID" json:"-"`
}

func (Bandul) TableName() string {
	return "bandul"
}