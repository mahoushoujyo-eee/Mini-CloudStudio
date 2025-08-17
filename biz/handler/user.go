package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"learn/biz/model"
	"learn/biz/service"
)

func UserRegister(ctx context.Context, c *app.RequestContext) {
	var err error
	var userParam model.UserParam

	err = c.BindAndValidate(&userParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	userId, err := service.NewUserService(ctx, c).Register(userParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Data:       userId,
		Message:    "注册成功",
	})
}

func UserHello(ctx context.Context, c *app.RequestContext) {
	usernameInterface, _ := c.Get("nickname")
	name, _ := usernameInterface.(string)
	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Data:       name,
		Message:    "Hello " + name,
	})
}

func UserResetCode(ctx context.Context, c *app.RequestContext) {
	var err error
	var emailParam model.EmailParam

	err = c.BindAndValidate(&emailParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = service.NewUserService(ctx, c).SendEmail(emailParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "Success",
	})
}

func UserResetPassword(ctx context.Context, c *app.RequestContext) {
	var err error
	var emailParam model.EmailParam

	err = c.BindAndValidate(&emailParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = service.NewUserService(ctx, c).ResetPassword(emailParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "Success",
	})
}
