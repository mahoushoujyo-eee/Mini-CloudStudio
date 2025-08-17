package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

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

func (s *AppService) CreateApp(appParam *model.AppParam) (string, error) {

	userId, ok := s.c.Get("user_id")

	if !ok {
		return "", errors.New("没有找到用户ID")
	}

	application := &model.Application{
		Name:   appParam.Name,
		UserId: uint(userId.(int64)),
		Cpu:    appParam.Cpu,
		Memory: appParam.Memory,
		State:  "initializing",
	}

	laterfix := uuid.NewString()[:8]

	kbParam := &model.KubernetesParam{
		Namespace: fmt.Sprintf("ns-%d", userId.(int64)),
		Pod:       fmt.Sprintf("pod-%s", laterfix),
		Svc:       fmt.Sprintf("svc-%s", laterfix),
		Pvc:       fmt.Sprintf("pvc-%s", laterfix),
		State:     "initializing",
	}

	// 确保命名空间存在
	if err := util.NewKubernetesUtil(s.ctx).EnsureNamespace(kbParam.Namespace); err != nil {
		log.Printf("创建命名空间失败: %v", err)
		return "", err
	}

	log.Printf("开始提交创建请求")

	go func() {
		util.NewKubernetesUtil(s.ctx).CreatePvc(kbParam, appParam)
		util.NewKubernetesUtil(s.ctx).CreatePod(kbParam, appParam)
		util.NewKubernetesUtil(s.ctx).CreateSvc(kbParam, appParam)
	}()

	err := config.DB.WithContext(s.ctx).Create(application).Error
	if err != nil {
		return "", err
	}

	return kbParam.Pod, nil
}

func (s *AppService) GetStateOfApp(kbParam *model.KubernetesParam) (*corev1.PodStatus, error) {
	userId, ok := s.c.Get("user_id")

	if !ok {
		return nil, errors.New("没有找到用户ID")
	}

	kbParam.Namespace = fmt.Sprintf("ns-%d", userId.(int64))

	podInfo, err := util.NewKubernetesUtil(s.ctx).GetPodInfo(kbParam)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return &podInfo.Status, nil
}

func (s *AppService) DeleteApp(appParam *model.AppParam) error {
	userId, ok := s.c.Get("user_id")

	if !ok {
		return errors.New("没有找到用户ID")
	}

	var laterfix string
	if len(appParam.Name) >= 8 {
		laterfix = appParam.Name[len(appParam.Name)-8:]
	} else {
		return errors.New("应用名称错误！")
	}

	kbParam := &model.KubernetesParam{
		Namespace: fmt.Sprintf("ns-%d", userId.(int64)),
		Pod:       fmt.Sprintf("pod-%s", laterfix),
		Svc:       fmt.Sprintf("svc-%s", laterfix),
		Pvc:       fmt.Sprintf("pvc-%s", laterfix),
	}

	err := util.NewKubernetesUtil(s.ctx).DeletePodSvcPvc(kbParam)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) ListApps(appParam *model.AppParam) ([]*model.Application, error) {
	var applications []*model.Application

	err := config.DB.Model(&model.Application{}).Where("user_id = ?", appParam.UserId).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}
