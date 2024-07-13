package test_helpers

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
	"user-service/internal/model"
)

var (
	usersCollection *mongo.Collection
	mongoClient     *mongo.Client
)

const (
	mongo_query_timeout      = time.Second
	mongo_disconnect_timeout = 2 * time.Second
)

func SetupMongoConnection() error {
	mongoOpts := options.Client().ApplyURI("mongodb://user:password@localhost:27017/").SetAppName("e2e-tests")
	var err error
	mongoClient, err = mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		return err
	}

	usersCollection = mongoClient.Database("demo").Collection("users")
	return nil
}

func CleanupMongoUsersCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongo_query_timeout)
	defer cancel()
	return usersCollection.Drop(ctx)
}

func CloseMongoConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongo_disconnect_timeout)
	defer cancel()
	return mongoClient.Disconnect(ctx)
}

func GetUserFromDB(t *testing.T, id uuid.UUID) model.User {
	ctx, cancel := context.WithTimeout(context.Background(), mongo_query_timeout)
	defer cancel()

	result := usersCollection.FindOne(ctx, bson.M{"_id": id})
	require.NoError(t, result.Err(), "failed to find user by ID")

	var user model.User
	err := result.Decode(&user)
	require.NoError(t, err, "failed to decode user")

	return user
}

func CreateUsersInDB(t *testing.T, user ...model.User) {
	for _, u := range user {
		CreateUserInDB(t, u)
	}
}

func CreateUserInDB(t *testing.T, user model.User) {
	ctx, cancel := context.WithTimeout(context.Background(), mongo_query_timeout)
	defer cancel()

	_, err := usersCollection.InsertOne(ctx, user)
	require.NoError(t, err, "failed to create user")
}

func AssertUsersDBCollectionIsEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), mongo_query_timeout)
	defer cancel()

	count, err := usersCollection.EstimatedDocumentCount(ctx)
	require.NoError(t, err, "failed to count users")

	assert.Empty(t, count, "expected to find no users")
}
