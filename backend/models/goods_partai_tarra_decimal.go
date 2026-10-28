package models

type GoodsPartaiDecimal struct {
	ID            uint    `gorm:"column:goods_partai_decimal_id;primaryKey;autoIncrement" json:"goods_partai_decimal_id"`
	IDGoodsPartai uint    `gorm:"column:goods_partai_id;not null;index" json:"goods_partai_id"`
	DecimalMin    float64 `gorm:"column:goods_partai_decimal_min;type:double precision;not null;default:0" json:"goods_partai_decimal_min"`
	DecimalMax    float64 `gorm:"column:goods_partai_decimal_max;type:double precision;not null;default:0" json:"goods_partai_decimal_max"`
	BrutoNominal  float64 `gorm:"column:goods_partai_bruto_nominal;type:double precision;not null;default:0" json:"goods_partai_bruto_nominal"`

	GoodsPartai *GoodsPartai `gorm:"foreignKey:IDGoodsPartai;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (GoodsPartaiDecimal) TableName() string {
	return "goods_partai_decimal"
}

type GoodsPartaiTarra struct {
	ID            uint    `gorm:"column:goods_partai_tarra_id;primaryKey;autoIncrement" json:"goods_partai_tarra_id"`
	IDGoodsPartai uint    `gorm:"column:goods_partai_id;not null;index" json:"goods_partai_id"`
	BeratMin      float64 `gorm:"column:goods_partai_tarra_berat_min;type:double precision;not null;default:0" json:"goods_partai_tarra_berat_min"`
	BeratMax      float64 `gorm:"column:goods_partai_tarra_berat_max;type:double precision;not null;default:0" json:"goods_partai_tarra_berat_max"`
	Nominal       float64 `gorm:"column:goods_partai_tarra_nominal;type:double precision;not null;default:0" json:"goods_partai_tarra_nominal"`

	GoodsPartai *GoodsPartai `gorm:"foreignKey:IDGoodsPartai;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (GoodsPartaiTarra) TableName() string {
	return "goods_partai_tarra"
}