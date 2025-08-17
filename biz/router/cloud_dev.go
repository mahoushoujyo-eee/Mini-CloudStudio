package router

import (
	"learn/biz/handler"
	"learn/biz/middleware"

	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterCloudDev(r *route.RouterGroup) {

	publicRouter := r.Group("/public")
	{
		publicRouter.GET("/")
	}

	commonRouter := r.Group("/common", middleware.JwtMiddleware.MiddlewareFunc())
	{
		commonRouter.GET("/hello", handler.UserHello)
		commonRouter.GET("/details")
		commonRouter.GET("/list")
		commonRouter.POST("/create", handler.AppCreate)
		commonRouter.POST("/delete")
		commonRouter.POST("/update")
	}

	adminRouter := r.Group("/admin")
	{
		adminRouter.GET("/")
	}
}
