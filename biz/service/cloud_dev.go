package service

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type AppService struct {
	ctx context.Context
	c   *app.RequestContext
}

func NewAppService(ctx context.Context, c *app.RequestContext) *AppService {
	return &AppService{ctx: ctx, c: c}
}
