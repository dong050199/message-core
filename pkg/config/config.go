package config

import (
	"log"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

var (
	kafkaConfig KafkaCfg
	redisClient RedisClientCfg
)

type KafkaCfg struct {
	InitTopics        bool   `envconfig:"KAFKA_INIT_TOPIC"`
	Brokers           string `envconfig:"KAFKA_BROKERS"`
	GroupID           string `envconfig:"KAFKA_GROUP_ID"`
	PoolSize          int    `envconfig:"KAFKA_POOL_SIZE"`
	Partition         int    `envconfig:"KAFKA_PATITION"`
	ReplicationFactor int    `envconfig:"KAFKA_REPLICATION"`
	// kafkaTopics...
	TopicDLQ           string `envconfig:"KAFKA_TOPIC_DLQ"`
	TopicBudgetProfile string `envconfig:"KAFKA_TOPIC_BUDGET_PROFILE"`
	// kafka retry opts...
	KafkaRetryAttempts     uint   `envconfig:"KAFKA_RETRY_ATTEMPTS"`
	KafkaRetryDelay        int    `envconfig:"KAFKA_RETRY_DELAYS"`
	PushFailedMessageToDLQ bool   `envconfig:"KAFKA_PUSH_FAILED_TO_DLQ"`
	DLQMessageKey          string `envconfig:"KAFKA_DLQ_MESSAGE_KEY"`
}

type RedisClientCfg struct {
	RedisURL       string `envconfig:"REDIS_URL"`
	RedisSigleMode bool   `envconfig:"REDIS_SIGLE_MODE"`
}

func (k *KafkaCfg) GetBrokers() []string {
	if len(kafkaConfig.Brokers) == 0 {
		return []string{}
	}
	return strings.Split(kafkaConfig.Brokers, ",")
}

func (k *KafkaCfg) GetTopicsConsume() []string {
	return []string{
		k.TopicBudgetProfile,
	}
}

func SetConfig() {
	configs := []interface{}{
		&kafkaConfig,
		&redisClient,
	}
	for _, instance := range configs {
		err := envconfig.Process("", instance)
		if err != nil {
			log.Fatalf("unable to init config: %v, err: %v", instance, err)
		}
	}
}

func KafkaConfig() KafkaCfg {
	return kafkaConfig
}

func RedisConfig() RedisClientCfg {
	return redisClient
}
