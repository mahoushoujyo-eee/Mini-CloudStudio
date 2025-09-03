package task

import (
	"context"
	"encoding/json"
	"learn/biz/model"
	"learn/biz/service"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

// KafkaConfig Kafka配置结构
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// 全局Kafka配置
var GlobalKafkaConfig *KafkaConfig

// 全局Kafka消费者客户端
var KafkaConsumer sarama.ConsumerGroup

// InitKafka 初始化Kafka配置和客户端
func InitKafka() {
	log.Print("初始化Kafka配置....")

	// 从环境变量获取配置
	brokers := []string{getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")}
	topic := getEnvOrDefault("KAFKA_TOPIC", "pod-heartbeat")
	groupID := getEnvOrDefault("KAFKA_GROUP_ID", "pod-heartbeat-consumer")

	GlobalKafkaConfig = &KafkaConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}

	log.Printf("Kafka配置初始化成功: brokers=%v, topic=%s, groupID=%s",
		GlobalKafkaConfig.Brokers, GlobalKafkaConfig.Topic, GlobalKafkaConfig.GroupID)

	// 创建全局Kafka消费者客户端
	cfg := GetDefaultConsumerConfig()
	client, err := sarama.NewConsumerGroup(GlobalKafkaConfig.Brokers, GlobalKafkaConfig.GroupID, cfg)
	if err != nil {
		log.Fatalf("创建Kafka消费者客户端失败: %v", err)
	}

	KafkaConsumer = client

	// 启动消费者
	go func() {
		for {
			// 消费消息
			err := client.Consume(context.Background(), []string{GlobalKafkaConfig.Topic}, &HeartbeatHandler{})
			if err != nil {
				log.Printf("Kafka消费者错误: %v", err)
			}
		}
	}()

	log.Print("Kafka消费者客户端初始化成功")
}

// GetDefaultConsumerConfig 获取默认的消费者配置
func GetDefaultConsumerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	cfg.Consumer.Group.Session.Timeout = 10 * time.Second
	cfg.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	return cfg
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ---------- 消息处理器 ----------
type HeartbeatHandler struct{}

func (h *HeartbeatHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *HeartbeatHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// 真正的消费逻辑
func (h *HeartbeatHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var record model.PodUsageRecord
		if err := json.Unmarshal(msg.Value, &record); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		if err := service.NewCounterService(context.TODO()).CountTime(record); err != nil {
			log.Printf("计数失败: %v", err)
		}

		log.Printf("收到心跳 | pod=%s namespace=%s user=%d lastUpdate=%s",
			record.PodName, record.Namespace, record.UserID, record.LastUpdate.Format(time.RFC3339))

		// 标记消息已处理
		sess.MarkMessage(msg, "")
	}
	return nil
}
