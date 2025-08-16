package model

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Type   string `gorm:"type:varchar(50);uniqueIndex" json:"name"`
	UserId uint   `gorm:"type:uint" json:"user_id"`
}
