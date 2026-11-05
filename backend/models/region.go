package models

// Province, City, dan Subdistrict = data wilayah yang sudah ada di database (dipakai bersama semua user).
// Model ini meniru tabel apa adanya. Kolom type:integer menjaga kolom int4 yang sudah ada
// agar tidak diubah AutoMigrate menjadi bigint.
//
// Nama field foreign key (IDProvince, IDCity) sengaja berbeda dari primary key tujuan (ID).
type Province struct {
	ID   uint   `gorm:"column:province_id;primaryKey;type:integer" json:"province_id"`
	Name string `gorm:"column:province;size:256" json:"province"`
}

func (Province) TableName() string {
	return "province"
}

type City struct {
	ID           uint   `gorm:"column:city_id;primaryKey;type:integer" json:"city_id"`
	IDProvince   uint   `gorm:"column:province_id;type:integer;index" json:"province_id"`
	ProvinceName string `gorm:"column:province;size:256" json:"province"`
	Type         string `gorm:"column:type;size:32" json:"type"`
	Name         string `gorm:"column:city_name;size:256" json:"city_name"`
	PostalCode   string `gorm:"column:postal_code;size:8" json:"postal_code"`
}

func (City) TableName() string {
	return "city"
}

type Subdistrict struct {
	ID           uint   `gorm:"column:subdistrict_id;primaryKey;type:integer" json:"subdistrict_id"`
	IDProvince   uint   `gorm:"column:province_id;type:integer;index" json:"province_id"`
	ProvinceName string `gorm:"column:province;size:256" json:"province"`
	IDCity       uint   `gorm:"column:city_id;type:integer;index" json:"city_id"`
	CityName     string `gorm:"column:city;size:256" json:"city"`
	Type         string `gorm:"column:type;size:32" json:"type"`
	Name         string `gorm:"column:subdistrict_name;size:256" json:"subdistrict_name"`
}

func (Subdistrict) TableName() string {
	return "subdistrict"
}