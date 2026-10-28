package models

import "time"

const (
	ItemTypeInternal = "internal"
	ItemTypeExternal = "external"
)

type BarcodeItem struct {
	ItemCode           string    `gorm:"column:item_code;primaryKey;size:32" json:"item_code"`
	ItemDateCreate     time.Time `gorm:"column:item_date_create;autoCreateTime" json:"item_date_create"`
	UserID             uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	ItemYear           int       `gorm:"column:item_year;not null;index:idx_barcode_item_scope,priority:3" json:"item_year"`
	ItemID             uint      `gorm:"column:item_id;not null;index" json:"item_id"`
	ItemCodeStatus     int       `gorm:"column:item_code_status;not null;default:1" json:"item_code_status"`
	IDGoodsPartai      uint      `gorm:"column:goods_partai_id;not null;index:idx_barcode_item_scope,priority:1" json:"goods_partai_id"`
	ItemType           string    `gorm:"column:item_type;size:32;not null;index:idx_barcode_item_scope,priority:2" json:"item_type"`
	ItemBarcodeType    *int      `gorm:"column:item_barcode_type" json:"item_barcode_type"`
	NamaRanting        *string   `gorm:"column:nama_ranting;size:255" json:"nama_ranting"`
	Code               *string   `gorm:"column:code;size:32" json:"code"`
	StatusUsedDjarumID *int      `gorm:"column:status_used_djarum_id" json:"status_used_djarum_id"`

	User        *User        `gorm:"foreignKey:UserID;references:ID" json:"-"`
	GoodsPartai *GoodsPartai `gorm:"foreignKey:IDGoodsPartai;references:ID" json:"-"`
}

func (BarcodeItem) TableName() string {
	return "barcode_item"
}