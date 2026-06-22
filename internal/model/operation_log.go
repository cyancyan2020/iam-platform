package model

import "time"

type OperationLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null"                  json:"user_id"`
	Username   string    `gorm:"type:varchar(64);not null" json:"username"`
	Method     string    `gorm:"type:varchar(10);not null" json:"method"`
	Path       string    `gorm:"type:varchar(256);not null" json:"path"`
	IP         string    `gorm:"type:varchar(45);not null" json:"ip"`
	UserAgent  string    `gorm:"type:varchar(512)"         json:"user_agent"`
	StatusCode int       `gorm:"not null"                  json:"status_code"`
	DurationMs int       `gorm:"not null"                  json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_log"
}
