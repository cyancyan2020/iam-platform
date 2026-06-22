package model

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	Name      string         `gorm:"type:varchar(64);not null" json:"name"`
	DataScope string         `gorm:"type:varchar(16);not null;default:self" json:"data_scope"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string {
	return "role"
}
