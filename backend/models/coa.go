package models

import "time"

type CoaType struct {
	CoaTypeID   uint      `gorm:"column:coa_type_id;primaryKey" json:"coa_type_id"`
	CoaTypeName string    `gorm:"column:coa_type_name;size:100;not null" json:"coa_type_name"`
	Coas        []Coa     `gorm:"foreignKey:CoaTypeID;references:CoaTypeID" json:"coas,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CoaType) TableName() string { return "coa_types" }

type Coa struct {
	CoaID     uint      `gorm:"column:coa_id;primaryKey" json:"coa_id"`
	CoaNo     string    `gorm:"column:coa_no;size:50;uniqueIndex;not null" json:"coa_no"`
	CoaName   string    `gorm:"column:coa_name;size:150;not null" json:"coa_name"`
	CoaTypeID uint      `gorm:"column:coa_type_id;not null" json:"coa_type_id"`
	CoaParent *uint     `gorm:"column:coa_parent" json:"coa_parent"`
	CoaHead   *bool     `gorm:"column:coa_head" json:"coa_head"`

	Type     *CoaType  `gorm:"foreignKey:CoaTypeID;references:CoaTypeID" json:"type,omitempty"`
	Parent   *Coa      `gorm:"foreignKey:CoaParent;references:CoaID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children []Coa     `gorm:"foreignKey:CoaParent;references:CoaID" json:"children,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Coa) TableName() string { return "coas" }