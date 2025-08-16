package service

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	"learn/biz/config"
	"learn/biz/model"
	"learn/biz/util"
)

type UserService struct {
	ctx context.Context
	c   *app.RequestContext
}

// NewUserService create user service
func NewUserService(ctx context.Context, c *app.RequestContext) *UserService {
	return &UserService{ctx: ctx, c: c}
}

func (s *UserService) Register(userParam model.UserParam) (uint, error) {
	var count int64
	if err := config.DB.Model(&model.User{}).Where("username = ?", userParam.Username).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("用户名已存在")
	}

	encodedPassword, _ := util.HashPassword(userParam.Password)

	user := model.User{
		Username: userParam.Username,
		Password: encodedPassword,
		Email:    userParam.Email,
		Nickname: userParam.Nickname,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return 0, err
	}

	return user.ID, nil
}
