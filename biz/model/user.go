package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Email    string `json:"email" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password string `json:"password" gorm:"type:varchar(255);not null"`
	Nickname string `json:"nickname" gorm:"type:varchar(50)"`
	Avatar   string `json:"avatar" gorm:"type:varchar(255)"`
}

type UserParam struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type EmailParam struct {
	Receiver string `json:"receiver"`
	Type     string `json:"type"`
}

type Response struct {
	StatusCode int         `json:"statuscode"`
	Data       interface{} `json:"data"`
	Message    string      `json:"message"`
}
