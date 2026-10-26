package models

import "time"

// CoaJournalType = tabel master tipe jurnal COA.
type CoaJournalType struct {
	IDTypeCoa      uint      `gorm:"column:id_type_coa;primaryKey;autoIncrement" json:"id_type_coa"`
	NameCoaJournal string    `gorm:"column:name_coa_journal;size:255;not null" json:"name_coa_journal"`
	Created        time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified       time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
}

func (CoaJournalType) TableName() string {
	return "coa_journal_type"
}

type CoaJournal struct {
	IDCoaAot uint      `gorm:"column:id_coa_aot;primaryKey;autoIncrement" json:"id_coa_aot"`
	TypeID   uint      `gorm:"column:id_type_coa;not null;uniqueIndex:uq_coa_journal_user_type,priority:2" json:"id_type_coa"`
	IDCoa    uint      `gorm:"column:coa_id;not null;index" json:"coa_id"`
	UserID   uint      `gorm:"column:user_id;not null;uniqueIndex:uq_coa_journal_user_type,priority:1" json:"user_id"`
	Created  time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`

	Type *CoaJournalType `gorm:"foreignKey:TypeID;references:IDTypeCoa" json:"type,omitempty"`
	Coa  *Coa            `gorm:"foreignKey:IDCoa;references:CoaID" json:"coa,omitempty"`
	User *User           `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (CoaJournal) TableName() string {
	return "coa_journal"
}