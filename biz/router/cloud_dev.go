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
		commonRouter.POST("/details", handler.AppGetPodInfo)
		commonRouter.GET("/list", handler.AppList)
		commonRouter.POST("/create", handler.AppCreate)
		commonRouter.POST("/stop", handler.AppStop)
		commonRouter.POST("/restart", handler.AppRestart)
		commonRouter.POST("/delete", handler.AppDelete)
		commonRouter.GET("/details/list", handler.AppGetPodStateList)
		commonRouter.POST("/log", handler.AppGetLog)
		commonRouter.POST("/update")
	}

	adminRouter := r.Group("/admin")
	{
		adminRouter.GET("/")
	}
}
