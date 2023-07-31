package kafka

import (
	"context"
	"message-core/pkg/config"

	log "github.com/sirupsen/logrus"

	"github.com/segmentio/kafka-go"
)

type KafkaWriter struct {
	KafkaWriter *kafka.Writer
}

var kafkaWriterSigleton *KafkaWriter

func InitKafkaProducer() {
	log := log.StandardLogger()
	kafkaCfg := config.KafkaConfig()
	brokers := kafkaCfg.GetBrokers()
	kafkaWriter := NewWriter(brokers, kafka.LoggerFunc(log.Errorf))
	if kafkaWriter == nil {
		log.Fatal("Failed to initialize kafka writer.")
	}
	kafkaWriterSigleton = &KafkaWriter{kafkaWriter}
}

func PublishMessage(ctx context.Context, msgs ...kafka.Message) error {
	log.WithField("KAFKA_PUBLISH_MESSAGE: ", msgs).WithContext(ctx).Info("Publish Kafka Message")
	return kafkaWriterSigleton.KafkaWriter.WriteMessages(ctx, msgs...)
}

func Close() error {
	return kafkaWriterSigleton.KafkaWriter.Close()
}
