package router

import (
	"learn/biz/handler"
	"learn/biz/middleware"

	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterUser(r *route.RouterGroup) {

	publicRouter := r.Group("/public")
	{
		publicRouter.POST("/login", middleware.JwtMiddleware.LoginHandler)
		publicRouter.POST("/register", handler.UserRegister)
		publicRouter.POST("/reset/email", handler.UserResetCode)
		publicRouter.POST("/reset/password", handler.UserResetPassword)
	}

	commonRouter := r.Group("/common", middleware.JwtMiddleware.MiddlewareFunc())
	{
		commonRouter.GET("/hello", handler.UserHello)
	}

	adminRouter := r.Group("/admin")
	{
		adminRouter.GET("/")
	}
}
