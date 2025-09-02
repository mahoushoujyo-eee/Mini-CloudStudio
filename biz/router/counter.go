package router

import (
	"learn/biz/handler"

	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterCounter(r *route.RouterGroup) {
	publicRouter := r.Group("/public")
	{
		publicRouter.POST("/time", handler.CountTime)
	}
}
