package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"

	"learn/biz/config"
	"learn/biz/model"
	"learn/biz/util"
)

type TimerService struct {
	ctx      context.Context
	cron     *cron.Cron
	redis    *redis.Client
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type PodUsageInfo struct {
	PodName      string    `json:"pod_name"`
	Namespace    string    `json:"namespace"`
	UserID       int64     `json:"user_id"`
	StartTime    time.Time `json:"start_time"`
	LastUpdate   time.Time `json:"last_update"`
	TotalSeconds int64     `json:"total_seconds"`
}

func NewTimerService(ctx context.Context) *TimerService {
	c := cron.New(cron.WithSeconds())
	return &TimerService{
		ctx:      ctx,
		cron:     c,
		redis:    config.RedisClient,
		stopChan: make(chan struct{}),
	}
}

func (s *TimerService) Start() {
	// 每30秒更新一次Pod使用时间
	_, err := s.cron.AddFunc("*/30 * * * * *", s.updatePodUsageTime)
	if err != nil {
		log.Fatalf("添加更新Pod使用时间任务失败: %v", err)
	}

	// 每5分钟同步一次数据到MySQL
	_, err = s.cron.AddFunc("0 */5 * * * *", s.syncToDatabase)
	if err != nil {
		log.Fatalf("添加同步数据任务失败: %v", err)
	}

	// 每天凌晨清理过期数据
	_, err = s.cron.AddFunc("0 0 0 * * *", s.cleanupExpiredData)
	if err != nil {
		log.Fatalf("添加清理过期数据任务失败: %v", err)
	}

	s.cron.Start()
	log.Println("计时服务已启动，计量单位：秒")
}

func (s *TimerService) Stop() {
	close(s.stopChan)
	s.cron.Stop()

	// 等待所有任务完成
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// 等待最多10秒
	select {
	case <-done:
		log.Println("所有任务已完成")
	case <-time.After(10 * time.Second):
		log.Println("等待任务完成超时")
	}

	// 最后同步一次数据
	s.syncToDatabase()
	log.Println("计时服务已安全停止")
}

// 更新Pod使用时间
func (s *TimerService) updatePodUsageTime() {
	s.wg.Add(1)
	defer s.wg.Done()

	select {
	case <-s.stopChan:
		return
	default:
	}

	// 获取所有用户的命名空间
	var applications []*model.Application
	err := config.DB.WithContext(s.ctx).Find(&applications).Error
	if err != nil {
		log.Printf("获取应用列表失败: %v", err)
		return
	}

	// 按用户ID分组
	userNamespaces := make(map[int64]string)
	for _, app := range applications {
		userNamespaces[int64(app.UserId)] = fmt.Sprintf("ns-%d", app.UserId)
	}

	// 遍历每个用户的命名空间
	for userID, namespace := range userNamespaces {
		select {
		case <-s.stopChan:
			return
		default:
			s.updateUserPods(userID, namespace)
		}
	}
}

func (s *TimerService) updateUserPods(userID int64, namespace string) {
	kbParam := &model.KubernetesParam{
		Namespace: namespace,
	}

	podList, err := util.NewKubernetesUtil(s.ctx).GetPodList(kbParam)
	if err != nil {
		log.Printf("获取Pod列表失败 - Namespace: %s, Error: %v", namespace, err)
		return
	}

	now := time.Now()

	for _, pod := range podList.Items {
		// 只处理运行中的Pod
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		redisKey := fmt.Sprintf("pod_usage:%s:%s", namespace, pod.Name)

		// 从Redis获取现有数据
		existingData, err := s.redis.Get(s.ctx, redisKey).Result()
		var usageInfo PodUsageInfo

		if errors.Is(err, redis.Nil) {
			// 新Pod，初始化数据
			usageInfo = PodUsageInfo{
				PodName:      pod.Name,
				Namespace:    namespace,
				UserID:       userID,
				StartTime:    pod.CreationTimestamp.Time,
				LastUpdate:   now,
				TotalSeconds: 0,
			}
		} else if err != nil {
			log.Printf("Redis获取数据失败: %v", err)
			continue
		} else {
			// 解析现有数据
			err = json.Unmarshal([]byte(existingData), &usageInfo)
			if err != nil {
				log.Printf("解析使用数据失败: %v", err)
				continue
			}
		}

		// 计算增量时间（秒）
		incrementSeconds := int64(now.Sub(usageInfo.LastUpdate).Seconds())
		usageInfo.TotalSeconds += incrementSeconds
		usageInfo.LastUpdate = now
		log.Printf("更新Pod使用时间 - Namespace: %s, Pod: %s, 增量时间: %d 秒", namespace, pod.Name, incrementSeconds)

		// 保存到Redis
		usageData, _ := json.Marshal(usageInfo)
		err = s.redis.Set(s.ctx, redisKey, usageData, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Redis保存数据失败: %v", err)
		}

		// 同时更新用户总使用时间的键
		userUsageKey := fmt.Sprintf("user_total_usage:%d", userID)
		s.redis.IncrBy(s.ctx, userUsageKey, incrementSeconds)
		s.redis.Expire(s.ctx, userUsageKey, 24*time.Hour)
	}
}

// 同步数据到MySQL
func (s *TimerService) syncToDatabase() {
	s.wg.Add(1)
	defer s.wg.Done()

	log.Println("开始同步pod使用数据到数据库")

	pattern := "pod_usage:*"
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		log.Printf("获取Redis键失败: %v", err)
		return
	}

	for _, key := range keys {
		select {
		case <-s.stopChan:
			return
		default:
		}

		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var usageInfo PodUsageInfo
		err = json.Unmarshal([]byte(data), &usageInfo)
		if err != nil {
			continue
		}

		// 更新或插入使用记录到数据库
		s.upsertUsageRecord(usageInfo)
	}

	log.Println("开始同步用户使用数据到数据库")
	pattern = "user_total_usage:*"
	keys, err = s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		log.Printf("获取Redis键失败: %v", err)
		return
	}

	for _, key := range keys {
		select {
		case <-s.stopChan:
			return
		default:
		}

		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		// Redis中存储的是累计的秒数
		totalSeconds, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			log.Printf("解析用户总使用时间失败: %v", err)
			continue
		}

		// 提取用户ID
		userIDStr := strings.TrimPrefix(key, "user_total_usage:")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			continue
		}

		if totalSeconds == 0 {
			continue
		}

		// 插入用户总使用记录到数据库
		s.upsertUserTotalUsage(userID, totalSeconds)

		// 清零Redis中的键（而不是删除，保持键存在以便继续累计）
		s.redis.Set(s.ctx, key, "0", 24*time.Hour)
	}
}

func (s *TimerService) upsertUsageRecord(usageInfo PodUsageInfo) {
	// 检查记录是否存在
	var existingRecord model.PodUsageRecord
	err := config.DB.WithContext(s.ctx).
		Where("pod_name = ? AND namespace = ? AND DATE(created_at) = ?",
			usageInfo.PodName, usageInfo.Namespace, time.Now().Format("2006-01-02")).
		First(&existingRecord).Error

	if err != nil {
		// 创建新记录
		record := &model.PodUsageRecord{
			PodName:      usageInfo.PodName,
			Namespace:    usageInfo.Namespace,
			UserID:       uint(usageInfo.UserID),
			StartTime:    usageInfo.StartTime,
			TotalSeconds: usageInfo.TotalSeconds,
			LastUpdate:   usageInfo.LastUpdate,
		}
		err = config.DB.WithContext(s.ctx).Create(record).Error
		if err != nil {
			log.Printf("创建Pod使用记录失败: %v", err)
		}
	} else {
		// 更新现有记录
		err = config.DB.WithContext(s.ctx).Model(&existingRecord).Updates(map[string]interface{}{
			"total_seconds": usageInfo.TotalSeconds,
			"last_update":   usageInfo.LastUpdate,
		}).Error
		if err != nil {
			log.Printf("更新Pod使用记录失败: %v", err)
		}
	}
}

// 同步用户总使用量到MySQL
func (s *TimerService) upsertUserTotalUsage(userID int64, totalSeconds int64) {
	// 创建新记录记录本次使用量
	record := &model.UserUsageRecord{
		UserID:       uint(userID),
		TotalSeconds: totalSeconds,
	}
	err := config.DB.WithContext(s.ctx).Create(record).Error
	if err != nil {
		log.Printf("创建用户总使用记录失败: %v", err)
	} else {
		log.Printf("成功记录用户 %d 的使用时间: %d 秒", userID, totalSeconds)
	}
}

// 清理过期数据
func (s *TimerService) cleanupExpiredData() {
	s.wg.Add(1)
	defer s.wg.Done()

	log.Println("开始清理过期数据")

	// 清理7天前的Redis数据
	pattern := "pod_usage:*"
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		return
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	for _, key := range keys {
		select {
		case <-s.stopChan:
			return
		default:
		}

		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var usageInfo PodUsageInfo
		err = json.Unmarshal([]byte(data), &usageInfo)
		if err != nil {
			continue
		}

		if usageInfo.LastUpdate.Before(sevenDaysAgo) {
			s.redis.Del(s.ctx, key)
			log.Printf("删除过期的Pod使用数据: %s", key)
		}
	}
}

// 获取用户的实时使用统计
func (s *TimerService) GetUserUsageStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取今日总使用时间
	userUsageKey := fmt.Sprintf("user_total_usage:%d", userID)
	todaySeconds, err := s.redis.Get(s.ctx, userUsageKey).Result()
	if errors.Is(err, redis.Nil) {
		stats["today_seconds"] = 0
	} else if err != nil {
		return nil, err
	} else {
		seconds, _ := strconv.ParseInt(todaySeconds, 10, 64)
		stats["today_seconds"] = seconds
	}

	// 获取活跃Pod数量
	namespace := fmt.Sprintf("ns-%d", userID)
	pattern := fmt.Sprintf("pod_usage:%s:*", namespace)
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	activePods := 0
	for _, key := range keys {
		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var usageInfo PodUsageInfo
		err = json.Unmarshal([]byte(data), &usageInfo)
		if err != nil {
			continue
		}

		// 如果最近5分钟有更新，认为是活跃的
		if time.Since(usageInfo.LastUpdate) < 5*time.Minute {
			activePods++
		}
	}

	stats["active_pods"] = activePods

	return stats, nil
}
