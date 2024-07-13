package events

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type KafkaProducer struct {
	p        *kafka.Producer
	eventsWG *sync.WaitGroup
}

// NewKafkaProducer connects to the Kafka bootstrap server, starts a goroutine that logs the received kafka events
// and returns a new KafkaProducer, that can be used to produce events to topics.
// To gracefully close the producer call Close().
func NewKafkaProducer(bootstrapServer string, opts ...KafkaConfigOption) (*KafkaProducer, error) {
	cfg := &kafka.ConfigMap{"bootstrap.servers": bootstrapServer}
	for _, opt := range opts {
		opt(cfg)
	}

	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create producer")
	}

	eventsWG := &sync.WaitGroup{}
	eventsWG.Add(1)
	go func() {
		defer eventsWG.Done()
		logEvents(p.Events())
	}()

	return &KafkaProducer{
		p:        p,
		eventsWG: eventsWG,
	}, nil
}

// Close gracefully closes the producer.
func (k *KafkaProducer) Close(flushTimeout time.Duration) {
	k.p.Flush(int(flushTimeout.Milliseconds()))
	k.p.Close()
	k.eventsWG.Wait()
}

// Produce produces given event data to the topic partition.
func (k *KafkaProducer) Produce(event []byte, tp kafka.TopicPartition) error {
	return k.p.Produce(&kafka.Message{
		TopicPartition: tp,
		Value:          event,
	}, nil)
}

// Health always reports the producer as healthy.
// Kafka go client lib is missing a support for checking health of kafka servers - no Ping() or similar func.
// We could be storing the (latest) failure kafka events and evaluate the health of kafka based on that - check if in the
// last xy seconds there was an error - if yes say it is currently unhealthy. Or something similar.
// Won't implement that for this homework project to keep it simple...
func (k *KafkaProducer) Health(_ context.Context) error {
	return nil
}

func logEvents(events chan kafka.Event) {
	// events channel is closed once we call Close() on the Producer
	for e := range events {
		switch ev := e.(type) {
		case kafka.Error:
			logrus.WithError(ev).WithFields(logrus.Fields{
				"retryable":  ev.IsRetriable(),
				"fatal":      ev.IsFatal(),
				"error_code": ev.Code(),
			}).Error("Kafka producer error")
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logrus.WithError(ev.TopicPartition.Error).
					Errorf("Failed to deliver message: %v", ev.TopicPartition)
			} else {
				logrus.Debugf("Successfully produced record to topic %s partition [%d] @ offset %v",
					*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
			}
		}
	}
}
