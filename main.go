package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hellofresh/health-go/v5"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	cfg "user-service/internal/configuration"
	"user-service/internal/controller"
	"user-service/internal/events"
	"user-service/internal/metrics"
	"user-service/internal/service"
	"user-service/internal/storage"
)

func main() {
	terminateChan := make(chan os.Signal, 1)
	defer signal.Stop(terminateChan)
	signal.Notify(terminateChan, syscall.SIGTERM, syscall.SIGINT)

	cfg, err := cfg.LoadFromEnvOrDefault()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load service config from environment")
	}
	metrics.RegisterHTTPMetrics()

	kafkaProducer, err := events.NewKafkaProducer(cfg.KafkaServer,
		events.WithAcks("all"),
		events.WithClientID(cfg.ServiceName),
		events.WithSecurityProtocol("plaintext"))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create kafka producer")
	}
	userEventsKafkaProducer := events.NewKafkaTopicProducer(kafkaProducer, cfg.KafkaEventsTopicName)

	mongoOpts := options.Client().ApplyURI(cfg.MongoURL).SetAppName(cfg.ServiceName)
	mongoClient, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to mongodb")
	}
	database := mongoClient.Database(cfg.MongoDBName)
	usersStore := storage.NewMongoUsersStorage(database, storage.WithTimeout(cfg.MongoOperationTimeout))

	healthHandler, err := createHealthHandler(cfg.ServiceName, mongoClient, kafkaProducer)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create health handler")
	}

	svc := service.New(usersStore, userEventsKafkaProducer)
	httpServer := setupHTTPServer(cfg.HTTPServerPort, svc, healthHandler.Handler())
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Fatal("failed to start HTTP server")
		}
	}()

	<-terminateChan
	logrus.Info("Shutting down service...")
	gracefulShutdown(cfg, httpServer, mongoClient, kafkaProducer)
	os.Exit(0)
}

func setupHTTPServer(port int, svc *service.Service, health http.Handler) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(metrics.HTTPRequestDurationMetricsMiddleware())
	router.Use(gin.LoggerWithWriter(logrus.StandardLogger().Out))

	v1Group := router.Group("v1")
	controller.CreateUsersHandlers(v1Group, svc)

	router.GET("/health", gin.WrapH(health))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router.Handler(),
	}
}

func createHealthHandler(serviceName string, mongo *mongo.Client, producer *events.KafkaProducer) (*health.Health, error) {
	return health.New(health.WithComponent(health.Component{
		Name:    serviceName,
		Version: "v1.0",
	}), health.WithChecks(health.Config{
		Name: "mongodb",
		Check: func(ctx context.Context) error {
			if err := mongo.Ping(ctx, readpref.Primary()); err != nil {
				return errors.Wrap(err, "mongoDB health check failed on ping")
			}
			return nil
		},
	},
		health.Config{
			Name:  "kafka",
			Check: producer.Health,
		}))
}

// gracefulShutdown at first shuts down the HTTP server, then mongo and kafka connections in parallel
func gracefulShutdown(cfg *cfg.ServiceConfig, server *http.Server, mongoClient *mongo.Client, kafkaProducer *events.KafkaProducer) {
	httpCtx, cancelHTTP := context.WithTimeout(context.Background(), cfg.HTTPGracefulShutdownTimeout)
	defer cancelHTTP()

	logrus.Info("Shutting down HTTP server")
	if err := server.Shutdown(httpCtx); err != nil {
		logrus.WithError(err).Fatal("Error while shutting down HTTP Server. Shutting down forcefully...")
	}

	mongoCtx, cancelMongo := context.WithTimeout(context.Background(), cfg.MongoGracefulShutdownTimeout)
	defer cancelMongo()
	var shutdownWG sync.WaitGroup
	shutdownWG.Add(1)
	go func() {
		logrus.Info("Disconnecting from mongo")
		defer shutdownWG.Done()
		if err := mongoClient.Disconnect(mongoCtx); err != nil {
			logrus.WithError(err).Fatal("Error while disconnecting from Mongo. Closing connection forcefully ...")
		}
	}()

	shutdownWG.Add(1)
	go func() {
		logrus.Info("Shutting down Kafka producer")
		defer shutdownWG.Done()
		kafkaProducer.Close(cfg.KafkaGracefulShutdownTimeout)
	}()

	shutdownWG.Wait()
}
