package events

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaTopicProducer struct {
	p              *KafkaProducer
	topicPartition kafka.TopicPartition
}

// NewKafkaTopicProducer creates new KafkaTopicProducer that produces events to given topic.
func NewKafkaTopicProducer(kp *KafkaProducer, topic string) *KafkaTopicProducer {
	return &KafkaTopicProducer{
		p:              kp,
		topicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
	}
}

// Produce marshals the given event into JSON and writes it to the kafka topic.
func (k *KafkaTopicProducer) Produce(event any) error {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return k.p.Produce(jsonBytes, k.topicPartition)
}
