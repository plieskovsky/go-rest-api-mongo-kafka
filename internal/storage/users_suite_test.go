package storage

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"runtime"
	"testing"
	"time"
	"user-service/internal/model"
)

type MongoTestSuite struct {
	suite.Suite
	server    *memongo.Server
	client    *mongo.Client
	db        *mongo.Database
	testStart time.Time
}

func Test_RunMongoTestSuite(t *testing.T) {
	suite.Run(t, new(MongoTestSuite))
}

// SetupSuite starts a mongo server for the tests to test the storage fully with the DB.
func (suite *MongoTestSuite) SetupSuite() {
	mongoServerOpts := &memongo.Options{
		MongoVersion:   "7.3.3",
		StartupTimeout: 15 * time.Second,
	}
	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		// Only set the custom url as workaround for arm64 macs (:
		mongoServerOpts.DownloadURL = "https://fastdl.mongodb.org/osx/mongodb-macos-arm64-7.3.3.tgz"
	}
	srv, err := memongo.StartWithOptions(mongoServerOpts)
	suite.Require().NoError(err, "failed to start mongoDB server")
	suite.server = srv

	ctx := context.Background()
	mongoOpts := options.Client().ApplyURI(srv.URI()).SetAppName("mongo-tests")
	client, err := mongo.Connect(ctx, mongoOpts)
	suite.Require().NoError(err, "Could not connect to Mongo")
	suite.client = client

	pingCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err = client.Ping(pingCtx, readpref.Primary())
	suite.Require().NoError(err, "Could not ping Mongo")

	suite.db = client.Database("test-database")
}

func (suite *MongoTestSuite) TearDownSuite() {
	disconnectCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := suite.client.Disconnect(disconnectCtx)
	suite.Assert().NoError(err, "Could not disconnect Mongo")

	suite.server.Stop()
}

func (suite *MongoTestSuite) BeforeTest(_, _ string) {
	// UTC & truncate to millis because that is the Mongo timezone & precision, so it can be used in assertions & test data
	suite.testStart = time.Now().UTC().Truncate(time.Millisecond)
}

func (suite *MongoTestSuite) createTestUsers(user ...model.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	for _, u := range user {
		_, err := suite.db.Collection("users").InsertOne(ctx, u)
		suite.Require().NoError(err, "creating test user")
	}
}
