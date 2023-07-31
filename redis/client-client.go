package redis

import (
	"context"
	"fmt"
	"log"
	"message-core/pkg/config"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
)

type RedisClient struct {
	Client *redis.Client
}

var redisClientSingleton *RedisClient

func InitRedisClient() {
	var ctx = context.Background()
	redisURL := config.RedisConfig().RedisURL
	fmt.Println(redisURL)
	logrus.WithField("Config redis", config.RedisConfig().RedisURL).Info()
	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set,
	})

	// Check heath
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logrus.WithField("Error when ping redis client", err).WithError(err).Error()
		return
	}

	client.AddHook(apmgoredis.NewHook())
	info, err := client.Info(ctx).Result()
	logrus.WithField("Redis client info", info).Info()
	if err != nil {
		logrus.WithField("Error when add hook redis", err).WithError(err).Error()
		return
	}
	redisClientSingleton = &RedisClient{Client: client}
}

func GetRedisClient() *redis.Client {
	if redisClientSingleton == nil {
		log.Fatal("failed to create new redis client")
	}

	return redisClientSingleton.Client
}

func Close() {
	redisClientSingleton.Client.Close()
}
