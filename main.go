package main

import (
	"log"
	"message-core/kafka"
	"message-core/mqtt"
	"message-core/pkg/config"
	"message-core/pkg/xservice/platform"
	"message-core/redis"
	"message-core/websocket"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	// init configuration
	config.InitConfig()

	// set configuration variables
	config.SetConfig()

	// init redis client
	redis.InitRedisClient()

	go InstanceWSserver()
	// create new connection to services
	platform.NewClien()

	// try to init config and kafka producer after making somethings noise
	kafka.InitKafkaProducer()

	// register
	mqtt.InstanceMQTTBroker()
}

func InstanceWSserver() {
	http.HandleFunc("/socket", websocket.HandleWS)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.WithField("Error when listening websocket server", err).WithError(err).Error()
		log.Fatal("Can't start server because websocket is not listening.")
	}
}
