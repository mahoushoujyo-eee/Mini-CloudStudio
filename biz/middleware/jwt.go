package middleware

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/jwt"

	"learn/biz/config"
	"learn/biz/model"
	"learn/biz/util"
)

var (
	JwtMiddleware *jwt.HertzJWTMiddleware
	identity      = "identity"
)

func InitJwt() {
	log.Printf("初始化 JWT...")

	var err error
	JwtMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Key:           []byte("tiktok secret key"),
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		Timeout:       5 * 24 * time.Hour,
		IdentityKey:   identity,
		// Verify password at login
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var userParam model.UserParam
			if err := c.BindAndValidate(&userParam); err != nil {
				return nil, err
			}

			var user model.User
			if err := config.DB.Where("email = ?", userParam.Email).Find(&user).Error; err != nil {
				return nil, err
			}

			result := util.CheckPassword(user.Password, userParam.Password)

			if !result {
				return nil, errors.New("密码不正确")
			}
			return &user, nil
		},
		// Set the payload in the token
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(*model.User); ok {
				user.Password = ""
				return jwt.MapClaims{
					"identity": user,
				}
			}
			return jwt.MapClaims{}
		},
		// build login response if verify password successfully
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			hlog.CtxInfof(ctx, "Login success ，token is issued clientIP: "+c.ClientIP())
			if code == 200 {
				// 登录成功，返回 token
				c.JSON(consts.StatusOK, model.Response{
					StatusCode: consts.StatusOK,
					Data:       token,
					Message:    "登录成功",
				})
			} else {
				// 登录失败，返回错误信息
				c.JSON(consts.StatusOK, model.Response{
					StatusCode: code,
					Message:    "登录失败",
				})
			}
		},
		// Verify token and get the id of logged-in user
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			// data 直接就是 identity 字段的值，即用户对象的 map 表示
			userMap, ok := data.(map[string]interface{})
			if !ok {
				log.Printf("无法转换为 map[string]interface{}")
				return false
			}

			// 从 map 中提取用户信息
			userID, ok := userMap["ID"].(float64) // JSON 中数字会被解析为 float64
			if !ok {
				log.Printf("无法获取用户 ID")
				return false
			}

			nickname, _ := userMap["nickname"].(string)
			email, _ := userMap["email"].(string)

			// 将用户信息存储到上下文中
			c.Set("user_id", int64(userID))
			c.Set("nickname", nickname)
			c.Set("email", email)

			hlog.CtxInfof(ctx, "Token is verified for user %d, clientIP: %s", int64(userID), c.ClientIP())
			return true
		},
		// Validation failed, build the message
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {

			log.Printf("未授权出错,获取到header:%v", string(c.GetHeader("Authorization")))

			c.JSON(consts.StatusOK, model.Response{
				StatusCode: code,
				Message:    "权限认证出错: " + message,
				Data:       nil,
			})
		},
		HTTPStatusMessageFunc: func(e error, ctx context.Context, c *app.RequestContext) string {
			if e != nil {
				return e.Error()
			}
			return "Unauthorized"
		},
	})

	if err != nil {
		panic(err)
	}
	log.Printf("初始化 JWT 完毕")
}
