package models

import "time"

// GoodsCategory = kategori barang (Tembakau, Non Tembakau, ...).
type GoodsCategory struct {
	ID        uint      `gorm:"column:goods_category_id;primaryKey;autoIncrement" json:"goods_category_id"`
	Name      string    `gorm:"column:goods_category_name;size:256;not null" json:"goods_category_name"`
	Desc      string    `gorm:"column:goods_category_desc;type:text" json:"goods_category_desc"`
	Active    int       `gorm:"column:goods_category_active;not null;default:1" json:"goods_category_active"`
	UserID    uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GoodsCategory) TableName() string {
	return "goods_category"
}

type Goods struct {
	ID              uint      `gorm:"column:goods_id;primaryKey;autoIncrement" json:"goods_id"`
	Name            string    `gorm:"column:goods_name;size:256;not null" json:"goods_name"`
	Desc            string    `gorm:"column:goods_desc;type:text" json:"goods_desc"`
	Active          int       `gorm:"column:goods_active;not null;default:1" json:"goods_active"`
	UserID          uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	IDGoodsCategory *uint     `gorm:"column:goods_category_id;index" json:"goods_category_id"`
	BrandID         *uint     `gorm:"column:goods_brand_id;index" json:"goods_brand_id"`
	UnitID          *uint     `gorm:"column:goods_unit_id;index" json:"goods_unit_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	User          *User          `gorm:"foreignKey:UserID;references:ID" json:"-"`
	GoodsCategory *GoodsCategory `gorm:"foreignKey:IDGoodsCategory;references:ID" json:"-"`
}

func (Goods) TableName() string {
	return "goods"
}

type GoodsCoa struct {
	ID         uint  `gorm:"column:goods_coa_id;primaryKey;autoIncrement" json:"goods_coa_id"`
	IDGoods    uint  `gorm:"column:goods_id;not null;uniqueIndex:idx_goods_coa_unique,priority:1" json:"goods_id"`
	IDCoaGoods uint  `gorm:"column:coa_goods_id;not null;uniqueIndex:idx_goods_coa_unique,priority:2" json:"coa_goods_id"`
	IDCoa      *uint `gorm:"column:coa_id;index" json:"coa_id"`
	UserID     uint  `gorm:"column:user_id;not null;index" json:"user_id"`

	Goods    *Goods    `gorm:"foreignKey:IDGoods;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CoaGoods *CoaGoods `gorm:"foreignKey:IDCoaGoods;references:ID" json:"-"`
	Coa      *Coa      `gorm:"foreignKey:IDCoa;references:CoaID" json:"coa,omitempty"`
	User     *User     `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GoodsCoa) TableName() string {
	return "goods_coa"
}