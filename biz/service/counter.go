package service

import (
	"context"
	"encoding/json"
	"errors"
	"learn/biz/config"
	"learn/biz/model"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8" // 添加这行
)

type CounterService struct {
	ctx context.Context
}

func NewCounterService(ctx context.Context) *CounterService {
	return &CounterService{ctx}
}

func (service *CounterService) CountTime(podUsage model.PodUsageRecord) error {
	//这里的podName要后续裁剪只留唯一ID！
	redisKey := "code-user" + strconv.Itoa(int(podUsage.UserID)) + ":" + podUsage.PodName
	result, err := config.RedisClient.Get(context.TODO(), redisKey).Result()

	// 正确处理 Redis 键不存在的情况
	if errors.Is(err, redis.Nil) {
		// 键不存在，创建新记录
		data, err := json.Marshal(podUsage)
		if err != nil {
			log.Printf("序列化数据失败 - Key: %s, Error: %v", redisKey, err)
			return err
		}
		err = config.RedisClient.Set(context.TODO(), redisKey, data, 0).Err()
		if err != nil {
			log.Printf("保存到Redis失败 - Key: %s, Error: %v", redisKey, err)
			return err
		}
		return nil
	} else if err != nil {
		// 其他错误
		log.Printf("从Redis获取数据失败 - Key: %s, Error: %v", redisKey, err)
		return err
	}

	// 键存在，解析现有数据
	var oldPodUsage model.PodUsageRecord
	err = json.Unmarshal([]byte(result), &oldPodUsage)
	if err != nil {
		log.Printf("解析Redis数据失败 - Key: %s, Error: %v", redisKey, err)
		return err
	}

	if oldPodUsage.StartTime != podUsage.StartTime {
		//如果启动时间不一致，就重新计算
		data, err := json.Marshal(podUsage)
		if err != nil {
			log.Printf("序列化数据失败 - Key: %s, Error: %v", redisKey, err)
			return err
		}
		err = config.RedisClient.Set(context.TODO(), redisKey, data, 0).Err()
		if err != nil {
			log.Printf("重新保存到Redis失败 - Key: %s, Error: %v", redisKey, err)
			return err
		}
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
		log.Printf("序列化更新数据失败 - Key: %s, Error: %v", redisKey, err)
		return err
	}
	err = config.RedisClient.Set(context.TODO(), redisKey, updatedData, 0).Err()
	if err != nil {
		log.Printf("更新Redis数据失败 - Key: %s, Error: %v", redisKey, err)
		return err
	}

	return nil
}
