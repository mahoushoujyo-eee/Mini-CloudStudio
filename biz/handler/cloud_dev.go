package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"learn/biz/model"
	"learn/biz/service"
)

func AppCreate(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam

	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = service.NewAppService(ctx, c).CreateApp(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "ok",
	})
}

func AppList(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam

	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	applications, err := service.NewAppService(ctx, c).ListApps(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "查询成功",
		Data:       applications,
	})

}
