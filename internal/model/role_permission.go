package model

import (
	"time"

	"gorm.io/gorm"
)

type RolePermission struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint64         `gorm:"not null" json:"role_id"`
	PermissionID uint64         `gorm:"not null" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}
