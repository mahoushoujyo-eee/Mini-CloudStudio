package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	corev1 "k8s.io/api/core/v1"

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

	podName, err := service.NewAppService(ctx, c).CreateApp(&appParam)
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
		Data:       podName,
	})
}

func AppStop(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam
	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	err = service.NewAppService(ctx, c).StopApp(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "ok",
	})

}

func AppRestart(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam
	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	err = service.NewAppService(ctx, c).RestartApp(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "ok",
	})
}

func AppGetPodInfo(ctx context.Context, c *app.RequestContext) {
	var kbParam model.KubernetesParam

	err := c.BindAndValidate(&kbParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	podInfo, err := service.NewAppService(ctx, c).GetStateOfApp(&kbParam)
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
		Data: struct {
			Phase      corev1.PodPhase       `json:"phase"`
			Conditions []corev1.PodCondition `json:"conditions"`
		}{
			Phase:      podInfo.Phase,
			Conditions: podInfo.Conditions,
		},
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

	applications, err := service.NewAppService(ctx, c).ListApp()
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

func AppDelete(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam

	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = service.NewAppService(ctx, c).DeleteApp(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, model.Response{
		StatusCode: consts.StatusOK,
		Message:    "删除成功",
	})
}

func AppGetPodStateList(ctx context.Context, c *app.RequestContext) {

	podList, err := service.NewAppService(ctx, c).GetPodStateList()
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
		Data:       podList,
	})
}

func AppGetLog(ctx context.Context, c *app.RequestContext) {
	var appParam model.AppParam

	err := c.BindAndValidate(&appParam)
	if err != nil {
		c.JSON(consts.StatusOK, model.Response{
			StatusCode: consts.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	logs, err := service.NewAppService(ctx, c).GetLogOfApp(&appParam)
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
		Data:       logs,
	})
}
