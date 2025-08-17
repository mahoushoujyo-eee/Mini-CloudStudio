package model

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	Name   string `gorm:"varchar(100); not null; unique" json:"name"`
	State  string `gorm:"varchar(20)" json:"state"`
	UserId uint   `gorm:"not null;" json:"user_id"`
	Cpu    int32  `gorm:"not null;" json:"cpu"`
	Memory uint64 `gorm:"not null;" json:"memory"`
}

type AppParam struct {
	Application
}

type KubernetesParam struct {
	Pod       string
	Namespace string
	State     string
	Pvc       string
	Pv        string
	Svc       string
}
