package kafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	log "github.com/sirupsen/logrus"
)

// MessageProcessor processor methods must implement kafka.Worker func method interface
type MessageProcessor interface {
	ProcessMessages(ctx context.Context, r *kafka.Reader, wg *sync.WaitGroup, workerID int)
}

// Worker kafka consumer worker fetch and process messages from reader
type Worker func(ctx context.Context, r *kafka.Reader, wg *sync.WaitGroup, workerID int)

type ConsumerGroup interface {
	ConsumeTopic(ctx context.Context, cancel context.CancelFunc, groupID, topic string, poolSize int, worker Worker)
	GetNewKafkaReader(kafkaURL []string, topic, groupID string) *kafka.Reader
	GetNewKafkaWriter(topic string) *kafka.Writer
}

type consumerGroup struct {
	Brokers []string
	GroupID string
	log     *log.Entry
}

// NewConsumerGroup kafka consumer group constructor
func NewConsumerGroup(brokers []string, groupID string, log *log.Entry) *consumerGroup {
	return &consumerGroup{Brokers: brokers, GroupID: groupID, log: log}
}

// GetNewKafkaReader create new kafka reader
func (c *consumerGroup) GetNewKafkaReader(kafkaURL []string, groupTopics []string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                kafkaURL,
		GroupID:                groupID,
		GroupTopics:            groupTopics,
		MinBytes:               minBytes,
		MaxBytes:               maxBytes,
		QueueCapacity:          queueCapacity,
		HeartbeatInterval:      heartbeatInterval,
		CommitInterval:         commitInterval,
		PartitionWatchInterval: partitionWatchInterval,
		MaxAttempts:            maxAttempts,
		MaxWait:                maxWait,
		Dialer: &kafka.Dialer{
			Timeout: dialTimeout,
		},
	})
}

// GetNewKafkaWriter create new kafka producer
func (c *consumerGroup) GetNewKafkaWriter() *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		MaxAttempts:  writerMaxAttempts,
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
	}

	return w
}

// ConsumeTopic start consumer group with given worker and pool size
func (c *consumerGroup) ConsumeTopic(groupTopics []string, poolSize int, worker Worker) {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer cancel()

	r := c.GetNewKafkaReader(c.Brokers, groupTopics, c.GroupID)

	defer func() {
		if err := r.Close(); err != nil {
			c.log.WithContext(ctx).WithField("consumerGroup.r.Close: ", err).Warn()
		}
	}()
	c.log.WithContext(ctx).
		WithField("consumerGroup",
			fmt.Sprintf("Starting consumer groupID: %s, topic: %+v, pool size: %v",
				c.GroupID,
				groupTopics,
				poolSize,
			)).Info()

	wg := &sync.WaitGroup{}
	for i := 0; i <= poolSize; i++ {
		wg.Add(1)
		go worker(ctx, r, wg, i)
	}
	wg.Wait()

	<-ctx.Done()
}
