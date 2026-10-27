package models

import "time"

// CoaGoods = master jenis akun untuk barang (Inventory, Sales, COGS, dst).
type CoaGoods struct {
	ID   uint   `gorm:"column:coa_goods_id;primaryKey;autoIncrement" json:"coa_goods_id"`
	Name string `gorm:"column:coa_goods_name;size:256;not null;uniqueIndex" json:"coa_goods_name"`
}

func (CoaGoods) TableName() string {
	return "coa_goods"
}

// GoodsPartai = data partai barang.
type GoodsPartai struct {
	ID          uint      `gorm:"column:goods_partai_id;primaryKey;autoIncrement" json:"goods_partai_id"`
	Code        int       `gorm:"column:goods_partai_code" json:"goods_partai_code"`
	Year        int       `gorm:"column:goods_partai_year" json:"goods_partai_year"`
	Name        string    `gorm:"column:goods_partai_name;size:256" json:"goods_partai_name"`
	Desc        string    `gorm:"column:goods_partai_desc;type:text" json:"goods_partai_desc"`
	Bungkus     string    `gorm:"column:goods_partai_bungkus;size:256" json:"goods_partai_bungkus"`
	Active      int       `gorm:"column:goods_partai_active" json:"goods_partai_active"`
	UserID      uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	RefGoodsID  *uint     `gorm:"column:ref_goods_id;index" json:"ref_goods_id"`
	AlertStatus int       `gorm:"column:goods_partai_alert_status;not null;default:0" json:"goods_partai_alert_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GoodsPartai) TableName() string {
	return "goods_partai"
}

type GoodsCoaPartai struct {
	ID            uint      `gorm:"column:goods_coa_partai_id;primaryKey;autoIncrement" json:"goods_coa_partai_id"`
	IDGoodsPartai uint      `gorm:"column:goods_partai_id;not null;uniqueIndex:uq_goods_coa_partai,priority:1" json:"goods_partai_id"`
	IDCoaGoods    uint      `gorm:"column:coa_goods_id;not null;uniqueIndex:uq_goods_coa_partai,priority:2" json:"coa_goods_id"`
	IDCoa         *uint     `gorm:"column:coa_id;index" json:"coa_id"`
	UserID        uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	GoodsPartai *GoodsPartai `gorm:"foreignKey:IDGoodsPartai;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CoaGoods    *CoaGoods    `gorm:"foreignKey:IDCoaGoods;references:ID" json:"coa_goods,omitempty"`
	Coa         *Coa         `gorm:"foreignKey:IDCoa;references:CoaID" json:"coa,omitempty"`
	User        *User        `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GoodsCoaPartai) TableName() string {
	return "goods_coa_partai"
}