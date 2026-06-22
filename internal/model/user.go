package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(64);default:''" json:"nickname"`
	RoleID       uint64         `gorm:"default:0" json:"role_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "user"
}
