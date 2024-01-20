package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id int `gorm:"primarykey"`

	CreatedAt  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	ModifiedAt time.Time `gorm:"type:TIMESTAMP;default:null" json:"modified_at"`
	DeletedAt  time.Time `gorm:"type:TIMESTAMP;default:null" json:"deleted_at"`

	CreatedBy  int            `gorm:"not null"`
	ModifiedBy *sql.NullInt64 `gorm:"null"`
	DeletedBy  *sql.NullInt64 `gorm:"null"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	value := tx.Statement.Context.Value("UserId")
	var userId = -1
	// TODO: check userId type
	if value != nil {
		userId = int(value.(float64))
	}
	m.CreatedAt = time.Now().UTC()
	m.CreatedBy = userId
	return
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	value := tx.Statement.Context.Value("UserId")
	var userId = &sql.NullInt64{Valid: false}
	// TODO: check userId type
	if value != nil {
		userId = &sql.NullInt64{Valid: true, Int64: int64(value.(float64))}
	}
	m.ModifiedAt = time.Now().UTC()
	m.ModifiedBy = userId
	return
}

func (m *BaseModel) BeforeDelete(tx *gorm.DB) (err error) {
	value := tx.Statement.Context.Value("UserId")
	var userId = &sql.NullInt64{Valid: false}
	// TODO: check userId type
	if value != nil {
		userId = &sql.NullInt64{Valid: true, Int64: int64(value.(float64))}
	}
	m.DeletedAt = time.Now().UTC()
	m.DeletedBy = userId
	return
}
