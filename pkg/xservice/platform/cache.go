package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"message-core/redis"
	"time"

	"github.com/sirupsen/logrus"
)

type UserCacheModel struct {
	UserState string         `json:"is_valid"`
	Rules     []RulesDevices `json:"rules"`
}

func SetUserCache(
	ctx context.Context,
	userName string,
	password string,
	req UserCacheModel,
) error {
	data, _ := json.Marshal(req)
	cmd := redis.GetRedisClient().Set(
		ctx,
		fmt.Sprintf("%s-%s", userName, password),
		data, 5*time.Minute)
	_, err := cmd.Result()
	if err != nil {
		return nil
	}
	return nil
}

// get cache not return error
func GetUserCache(
	ctx context.Context,
	userName string,
	password string,
) (resp UserCacheModel, err error) {
	cmd := redis.GetRedisClient().Get(
		ctx,
		fmt.Sprintf("%s-%s", userName, password))
	data, err := cmd.Result()
	if err != nil {
		logrus.WithError(err).
			WithField("GetRedisClient",
				fmt.Sprintf("%s-%s", userName, password)).Error()
		return resp, nil
	}
	if len(data) == 0 {
		return
	}
	err = json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return resp, nil
	}

	return
}

func SetRuleCache(
	ctx context.Context,
	userName string,
	req UserCacheModel,
) error {
	data, _ := json.Marshal(req)
	cmd := redis.GetRedisClient().Set(
		ctx,
		userName,
		data, 5*time.Minute)
	_, err := cmd.Result()
	if err != nil {
		return nil
	}
	return nil
}

// get cache not return error
func GetRuleCache(
	ctx context.Context,
	userName string,
) (resp UserCacheModel, err error) {
	cmd := redis.GetRedisClient().Get(
		ctx,
		userName,
	)
	data, err := cmd.Result()
	if err != nil {
		logrus.WithError(err).
			WithField("GetRedisClient", userName).Error()
		return resp, nil
	}
	if len(data) == 0 {
		return
	}
	err = json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return resp, nil
	}

	return
}
