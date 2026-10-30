package models

import "time"

// GoodsGrade = grade barang beserta harga per grade.
type GoodsGrade struct {
	ID        uint      `gorm:"column:goods_grade_id;primaryKey;autoIncrement" json:"goods_grade_id"`
	Name      string    `gorm:"column:goods_grade_name;size:256;not null" json:"goods_grade_name"`
	Desc      string    `gorm:"column:goods_grade_desc;type:text" json:"goods_grade_desc"`
	Price     float64   `gorm:"column:goods_grade_price;not null;default:0" json:"goods_grade_price"`
	Active    int       `gorm:"column:goods_grade_active;not null;default:1" json:"goods_grade_active"`
	UserID    uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GoodsGrade) TableName() string {
	return "goods_grade"
}