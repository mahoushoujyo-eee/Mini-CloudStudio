package service

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	"learn/biz/config"
	"learn/biz/model"
)

type AppService struct {
	ctx context.Context
	c   *app.RequestContext
}

func NewAppService(ctx context.Context, c *app.RequestContext) *AppService {
	return &AppService{ctx: ctx, c: c}
}

func (s *AppService) ListPods() ([]*model.Application, error) {
	userId, ok := s.c.Get("user_id")
	if !ok {
		return nil, errors.New("没有找到用户ID")
	}

	var applications []*model.Application

	err := config.DB.WithContext(s.ctx).
		Where("user_id = ?", userId).
		Find(&applications).Error

	if err != nil {
		return nil, err
	}

	return applications, nil
}
