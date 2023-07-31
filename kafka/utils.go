package kafka

import (
	"context"
	"message-core/pkg/config"

	log "github.com/sirupsen/logrus"

	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

func ConnectKafkaBrokers(ctx context.Context, cfg config.KafkaCfg) (conn *kafka.Conn, err error) {
	conn, err = NewKafkaConn(ctx, &Config{
		Brokers:    cfg.GetBrokers(),
		GroupID:    cfg.GroupID,
		InitTopics: cfg.InitTopics,
	}) // TODO:
	if err != nil {
		return nil, errors.Wrap(err, "kafka.NewKafkaCon")
	}

	brokers, err := conn.Brokers()
	if err != nil {
		return nil, errors.Wrap(err, "kafkaConn.Brokers")
	}

	log.WithField("KAFKA CONNECTED BROKERS: ", brokers).WithContext(ctx).Info("Kafka connected")

	return
}

func InitKafkaTopics(ctx context.Context, kafkaConn *kafka.Conn, topics ...kafka.TopicConfig) {

	controller, err := kafkaConn.Controller()
	if err != nil {
		log.Warnf("kafkaConn.Controller err: %v", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	log.Infof("kafka controller uri: %s", controllerURI)

	conn, err := kafka.DialContext(ctx, "tcp", controllerURI)
	if err != nil {
		log.Warnf("initKafkaTopics.DialContext: %v", err)
		return
	}
	defer conn.Close() //nolint: errcheck

	log.Infof("established new kafka controller connection: %s", controllerURI)

	if err := conn.CreateTopics(
		topics...,
	); err != nil {
		log.Warnf("kafkaConn.CreateTopics: %v", err)
		return
	}
	log.WithField("KAFKA TOPICS CREATED OR ALREADY EXISTS: ", topics).WithContext(ctx).Info("Kafka auto create topics")
}

func LogProcessMessage(ctx context.Context, m kafka.Message, workerID int) {
	log.WithFields(log.Fields{
		"Topic":     m.Topic,
		"Partition": m.Partition,
		"WorkerID":  workerID,
		"Offset":    m.Offset,
		"Time":      m.Time,
		"Message":   m.Value,
	}).WithContext(ctx).Info()
}
