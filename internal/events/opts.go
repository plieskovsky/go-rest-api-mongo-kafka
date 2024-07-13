package events

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfigOption func(configMap *kafka.ConfigMap)

func WithSecurityProtocol(securityProtocol string) KafkaConfigOption {
	return WithOption("security.protocol", securityProtocol)
}

func WithAcks(acks string) KafkaConfigOption {
	return WithOption("acks", acks)
}

func WithClientID(clientID string) KafkaConfigOption {
	return WithOption("client.id", clientID)
}

func WithOption(key, value string) KafkaConfigOption {
	return func(configMap *kafka.ConfigMap) {
		// ignore error as it is always nil
		_ = configMap.SetKey(key, value)
	}
}
