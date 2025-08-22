package model

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	Name       string `gorm:"type:varchar(100); not null;" json:"name"`
	PodName    string `gorm:"type:varchar(100); not null; unique" json:"pod_name"`
	UserId     uint   `gorm:"type:integer; not null;" json:"user_id"`
	Cpu        string `gorm:"type:varchar(100); not null;" json:"cpu"`
	Memory     string `gorm:"type:varchar(100); not null;" json:"memory"`
	Url        string `gorm:"type:varchar(255); not null;" json:"url"`
	Deployment string `gorm:"type:varchar(100); not null;" json:"deployment"`
	State      string `gorm:"-" json:"state"`
}

type AppParam struct {
	Application
	PodPassword string `json:"pod_password"`
}

type KubernetesParam struct {
	Deployment string
	Pod        string
	Namespace  string
	State      string
	Pvc        string
	Svc        string
	Cpu        string
	Memory     string
}
