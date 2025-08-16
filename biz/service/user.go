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
	if err := config.DB.Model(&model.User{}).Where("email = ?", userParam.Email).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("邮箱已被注册")
	}

	if err := config.DB.Model(&model.User{}).Where("username = ?", userParam.Username).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("用户名已存在")
	}

	code, err := config.RedisClient.Get(s.ctx, userParam.Email+":register").Result()
	if err == nil && code != "" {
		return 0, errors.New("确认验证码失败，请稍后再试")
	}

	if code != userParam.Code {
		return 0, errors.New("验证码错误")
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

	if err := config.DB.Create(&model.Role{UserId: user.ID, Type: "user"}).Error; err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (s *UserService) SendEmail(emailParam model.EmailParam) error {
	if emailParam.Receiver == "" || emailParam.Type == "" {
		return errors.New("参数部分为空")
	}

	code, err := config.Send(emailParam)
	if err != nil {
		return err
	}

	err = config.RedisClient.Set(s.ctx, emailParam.Receiver+":"+emailParam.Type, code, 5*60).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) ResetPassword(emailParam model.EmailParam) error {

	code, err := config.RedisClient.Get(s.ctx, emailParam.Receiver+":reset").Result()
	if err != nil {
		return err
	}

	if emailParam.Code != code {
		return errors.New("验证码错误")
	}

	encryptedPassword, err := util.HashPassword(emailParam.Password)

	if err != nil {
		return err
	}
	err = config.DB.Model(&model.User{}).Where("email = ?", emailParam.Receiver).Update("password", encryptedPassword).Error

	if err != nil {
		return err
	}

	return nil
}
