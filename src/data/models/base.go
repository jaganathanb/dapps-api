package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id int `gorm:"primarykey"`

	CreatedAt  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null" json:"createdAt"`
	ModifiedAt time.Time `gorm:"type:TIMESTAMP;default:null"                       json:"modifiedAt"`
	DeletedAt  time.Time `gorm:"type:TIMESTAMP;default:null"                       json:"deletedAt"`

	CreatedBy  int            `gorm:"not null" json:"craetedBy"`
	ModifiedBy *sql.NullInt64 `gorm:"null" json:"modifiedBy"`
	DeletedBy  *sql.NullInt64 `gorm:"null" json:"deletedBy"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedAt = time.Now().UTC()
	return
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	m.ModifiedAt = time.Now().UTC()
	return
}

func (m *BaseModel) BeforeDelete(tx *gorm.DB) (err error) {
	m.DeletedAt = time.Now().UTC()
	return
}
