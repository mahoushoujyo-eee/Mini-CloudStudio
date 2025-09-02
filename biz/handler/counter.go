package handler

import (
	"context"
	"learn/biz/model"
	"learn/biz/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func CountTime(ctx context.Context, c *app.RequestContext) {
	var podUsage model.PodUsageRecord
	err := c.BindAndValidate(&podUsage)
	if err != nil {
		c.JSON(200, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = service.NewCounterService(ctx, c).CountTime(podUsage)
	if err != nil {
		c.JSON(200, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(200, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "success",
	})
}
