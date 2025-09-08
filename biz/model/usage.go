package model

import (
	"time"

	"gorm.io/gorm"
)

type PodUsageRecord struct {
	gorm.Model
	PodName      string    `gorm:"not null;type:varchar(191);uniqueIndex:uniq_pod_ns_user" json:"pod_name"`
	Namespace    string    `gorm:"not null;type:varchar(191);uniqueIndex:uniq_pod_ns_user" json:"namespace"`
	UserID       uint      `gorm:"not null;uniqueIndex:uniq_pod_ns_user" json:"user_id"`
	StartTime    time.Time `gorm:"not null" json:"start_time"`
	TotalSeconds int64     `gorm:"default:0" json:"total_seconds"`
	LastUpdate   time.Time `gorm:"not null" json:"last_update"`
}

type UserUsageRecord struct {
	gorm.Model
	UserID       uint  `gorm:"not null;index" json:"user_id"`
	TotalSeconds int64 `gorm:"default:0" json:"total_seconds"`
}
