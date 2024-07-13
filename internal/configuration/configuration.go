package configuration

import (
	"os"
	"strconv"
	"time"
)

const (
	// keys
	http_server_port_key               = "HTTP_PORT"
	http_graceful_shutdown_period_key  = "HTTP_GRACEFUL_SHUTDOWN_PERIOD"
	mongo_graceful_shutdown_period_key = "MONGO_GRACEFUL_SHUTDOWN_PERIOD"
	kafka_graceful_shutdown_period_key = "KAFKA_GRACEFUL_SHUTDOWN_PERIOD"
	mongo_operation_timeout_key        = "MONGO_OPERATION_TIMEOUT"
	mongo_url_key                      = "MONGO_URL"
	mongo_db_name_key                  = "MONGO_DB_NAME"
	kafka_server_key                   = "KAFKA_SERVER"
	kafka_events_topic_name_key        = "EVENTS_TOPIC_NAME"

	// default values
	http_server_port_default               = 8080
	http_graceful_shutdown_period_default  = 5 * time.Second
	mongo_graceful_shutdown_period_default = 5 * time.Second
	kafka_graceful_shutdown_period_default = 5 * time.Second
	mongo_operation_timeout_default        = 3 * time.Second
	mongo_url_default                      = "mongodb://user:password@localhost:27017/"
	mongo_db_name_default                  = "demo"
	kafka_server_default                   = "localhost:9092"
	kafka_events_topic_name_default        = "UserEvents"
)

type ServiceConfig struct {
	ServiceName                  string
	HTTPServerPort               int
	HTTPGracefulShutdownTimeout  time.Duration
	MongoGracefulShutdownTimeout time.Duration
	KafkaGracefulShutdownTimeout time.Duration
	MongoOperationTimeout        time.Duration
	MongoURL                     string
	MongoDBName                  string
	KafkaServer                  string
	KafkaEventsTopicName         string
}

// LoadFromEnvOrDefault loads the service configuration variables from environment or sets them to default if not present.
// Error is returned when some environment variable parsing fails.
func LoadFromEnvOrDefault() (*ServiceConfig, error) {
	cfg := &ServiceConfig{
		ServiceName: "user-service",
	}

	// numeric ones
	num, err := getEnvOrDefaultInt(http_server_port_key, http_server_port_default)
	if err != nil {
		return nil, err
	}
	cfg.HTTPServerPort = *num

	//duration ones
	for durationCfgVar, varSettings := range map[*time.Duration]struct {
		key    string
		defVal time.Duration
	}{
		&cfg.MongoOperationTimeout:        {key: mongo_operation_timeout_key, defVal: mongo_operation_timeout_default},
		&cfg.KafkaGracefulShutdownTimeout: {key: kafka_graceful_shutdown_period_key, defVal: kafka_graceful_shutdown_period_default},
		&cfg.MongoGracefulShutdownTimeout: {key: mongo_graceful_shutdown_period_key, defVal: mongo_graceful_shutdown_period_default},
		&cfg.HTTPGracefulShutdownTimeout:  {key: http_graceful_shutdown_period_key, defVal: http_graceful_shutdown_period_default},
	} {
		dur, err := getEnvOrDefaultDuration(varSettings.key, varSettings.defVal)
		if err != nil {
			return nil, err
		}
		*durationCfgVar = *dur
	}

	// string ones
	cfg.KafkaServer = getEnvOrDefaultString(kafka_server_key, kafka_server_default)
	cfg.KafkaEventsTopicName = getEnvOrDefaultString(kafka_events_topic_name_key, kafka_events_topic_name_default)
	cfg.MongoURL = getEnvOrDefaultString(mongo_url_key, mongo_url_default)
	cfg.MongoDBName = getEnvOrDefaultString(mongo_db_name_key, mongo_db_name_default)

	return cfg, nil
}

func getEnvOrDefaultString(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvOrDefaultInt(key string, def int) (*int, error) {
	return getEnvOrDefault(key, def, strconv.Atoi)
}

func getEnvOrDefaultDuration(key string, def time.Duration) (*time.Duration, error) {
	return getEnvOrDefault(key, def, time.ParseDuration)
}

func getEnvOrDefault[T any](key string, def T, mapFunc func(string) (T, error)) (*T, error) {
	v := os.Getenv(key)
	if v == "" {
		return &def, nil
	}

	parsed, err := mapFunc(v)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
