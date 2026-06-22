package model

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Path      string         `gorm:"type:varchar(128);not null" json:"path"`
	Method    string         `gorm:"type:varchar(10);not null" json:"method"`
	Name      string         `gorm:"type:varchar(64);not null" json:"name"`
	ParentID  uint64         `gorm:"default:0" json:"parent_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Permission) TableName() string {
	return "permission"
}
