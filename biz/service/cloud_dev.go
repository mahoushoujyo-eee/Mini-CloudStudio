package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	"learn/biz/config"
	"learn/biz/model"
	"learn/biz/util"
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

func (s *AppService) CreateApp(appParam *model.AppParam) error {

	userId, ok := s.c.Get("user_id")

	if !ok {
		return errors.New("没有找到用户ID")
	}

	application := &model.Application{
		Name:   appParam.Name,
		UserId: userId.(uint),
		Cpu:    appParam.Cpu,
		Memory: appParam.Memory,
		State:  "initializing",
	}

	laterfix := uuid.NewString()[:8]

	kbParam := &model.KubernetesParam{
		Namespace: fmt.Sprintf("ns-%d", userId.(uint)),
		Pod:       fmt.Sprintf("pod-%s", laterfix),
		Svc:       fmt.Sprintf("svc-%s", laterfix),
		Pvc:       fmt.Sprintf("pvc-%s", laterfix),
		State:     "initializing",
	}

	go func() {
		util.NewKubernetesUtil(s.ctx).CreatePvc(kbParam, appParam)
		util.NewKubernetesUtil(s.ctx).CreatePod(kbParam, appParam)
		util.NewKubernetesUtil(s.ctx).CreateSvc(kbParam, appParam)
	}()

	err := config.DB.WithContext(s.ctx).Create(application).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) DeleteApp(appParam *model.AppParam) error {
	return nil
}

func (s *AppService) GetStateOfApp(appParam *model.AppParam) error {
	return nil
}
