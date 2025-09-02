package service

import (
	"context"
	"encoding/json"
	"learn/biz/config"
	"learn/biz/model"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
)

type CounterService struct {
	ctx context.Context
	c   *app.RequestContext
}

func NewCounterService(ctx context.Context, c *app.RequestContext) *CounterService {
	return &CounterService{ctx, c}
}

func (service *CounterService) CountTime(podUsage model.PodUsageRecord) error {
	//这里的podName要后续裁剪只留唯一ID！
	redisKey := "user" + strconv.Itoa(int(podUsage.UserID)) + ":" + podUsage.PodName
	result, err := config.RedisClient.Get(context.TODO(), redisKey).Result()
	if err != nil {
		return err
	}
	//如果不存在，就创建一个  Value待修改
	if result == "" {
		config.RedisClient.Set(context.TODO(), redisKey, podUsage, 0)
	} else {
		var oldPodUsage model.PodUsageRecord
		err = json.Unmarshal([]byte(result), &oldPodUsage)
		if err != nil {
			return err
		}

		if oldPodUsage.StartTime != podUsage.StartTime {
			//如果启动时间不一致，就重新计算
			config.RedisClient.Set(context.TODO(), redisKey, podUsage, 0)
			return nil
		}
		//如果存在，就更新
		//更新使用时间
		timeDiff := podUsage.LastUpdate.Sub(oldPodUsage.LastUpdate)
		oldPodUsage.TotalSeconds += int64(timeDiff.Seconds())

		// 更新最后使用时间
		oldPodUsage.LastUpdate = podUsage.LastUpdate

		// 将更新后的数据保存回Redis
		updatedData, err := json.Marshal(oldPodUsage)
		if err != nil {
			return err
		}
		config.RedisClient.Set(context.TODO(), redisKey, updatedData, 0)
	}
	return nil
}
