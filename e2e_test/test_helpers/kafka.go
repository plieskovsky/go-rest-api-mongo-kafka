package test_helpers

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"user-service/internal/model"
)

// if event was written into the topic it will be consumed sooner than in 3 seconds
const kafka_read_timeout = 3 * time.Second

var kafkaConsumer *kafka.Consumer

type CreateUpdateUserEvent struct {
	Action   string     `json:"action"`
	UserData model.User `json:"user_data"`
}

type DeleteUserEvent struct {
	Action   string    `json:"action"`
	UserData DeletedID `json:"user_data"`
}

type DeletedID struct {
	ID string `json:"id"`
}

func SetupKafkaConsumer() error {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
		"group.id":          "test-consumer",
		"auto.offset.reset": "smallest"})

	if err != nil {
		return err
	}

	err = consumer.SubscribeTopics([]string{"UserEvents"}, nil)
	if err != nil {
		return err
	}

	kafkaConsumer = consumer
	return nil
}

func CloseKafkaConsumer() error {
	err := kafkaConsumer.Unsubscribe()
	if err != nil {
		return err
	}
	return kafkaConsumer.Close()
}

func GetKafkaCreateOrUpdateEvent(t *testing.T) CreateUpdateUserEvent {
	return getKafkaEvent[CreateUpdateUserEvent](t)
}

func GetKafkaDeletedEvent(t *testing.T) DeleteUserEvent {
	return getKafkaEvent[DeleteUserEvent](t)
}

func getKafkaEvent[T any](t *testing.T) T {
	msg, err := kafkaConsumer.ReadMessage(kafka_read_timeout)
	require.NoError(t, err, "failed to read message")

	var event T
	err = json.Unmarshal(msg.Value, &event)
	require.NoError(t, err, "failed to unmarshal event")

	return event
}

func AssertNoUserEventPublishedToKafka(t *testing.T) {
	_, err := kafkaConsumer.ReadMessage(kafka_read_timeout)

	require.Error(t, err, "expected to receive error")
	require.Equal(t, kafka.ErrTimedOut, err.(kafka.Error).Code())
}
