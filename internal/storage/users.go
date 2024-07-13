package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"user-service/internal/model"
)

const defaultDBTimeout = 1 * time.Second

type Opt func(*MongoUsersStorage)

func WithTimeout(timeout time.Duration) Opt {
	return func(s *MongoUsersStorage) {
		s.dbTimeout = timeout
	}
}

type MongoUsersStorage struct {
	users     *mongo.Collection
	dbTimeout time.Duration
}

// NewMongoUsersStorage creates new storage that manages "users" collection in the given db.
func NewMongoUsersStorage(db *mongo.Database, opts ...Opt) *MongoUsersStorage {
	m := &MongoUsersStorage{
		users:     db.Collection("users"),
		dbTimeout: defaultDBTimeout,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// CreateUser creates the user in the DB. If DB operation fails the unchanged error is returned.
func (m MongoUsersStorage) CreateUser(ctx context.Context, user model.User) error {
	var dbCtx, cancel = context.WithTimeout(ctx, m.dbTimeout)
	defer cancel()

	_, err := m.users.InsertOne(dbCtx, user)
	if err != nil {
		return err
	}

	return nil
}

// GetUser gets the user from the DB based on the provided id. If no user is found NotFoundError error is returned.
// If DB operation fails the unchanged error is returned.
func (m MongoUsersStorage) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var dbCtx, cancel = context.WithTimeout(ctx, m.dbTimeout)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$eq": id}}
	result := m.users.FindOne(dbCtx, filter)
	if err := result.Err(); err != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, NotFoundError
		}
		return nil, err
	}

	var user model.User
	err := result.Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUsers fetches User slice from the DB. Sort field has to be set in the given params.
// If DB operation fails the unchanged error is returned.
func (m MongoUsersStorage) GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error) {
	var dbCtx, cancel = context.WithTimeout(ctx, m.dbTimeout)
	defer cancel()

	opts, err := createGetUsersOpts(params)
	if err != nil {
		return nil, err
	}
	filter := createGetUsersFilter(params)

	cursor, err := m.users.Find(dbCtx, filter, opts)
	if err != nil {
		return nil, err
	}

	var users []model.User
	if err = cursor.All(dbCtx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates the user in the DB while ignoring the created_at field. Returns the updated user.
// If the user is not found NotFoundError is returned.
// If the DB response data fails to be unmarshalled ResponseUnmarshallError is returned.
// If DB operation fails the unchanged error is returned.
func (m MongoUsersStorage) UpdateUser(ctx context.Context, user model.User) (*model.User, error) {
	var dbCtx, cancel = context.WithTimeout(ctx, m.dbTimeout)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$eq": user.ID}}
	update := bson.M{
		"$set": bson.M{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"nickname":   user.Nickname,
			"password":   user.Password,
			"email":      user.Email,
			"country":    user.Country,
			"updated_at": user.UpdatedAt,
		},
	}

	result := m.users.FindOneAndUpdate(dbCtx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err := result.Err(); err != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, NotFoundError
		}
		return nil, err
	}

	var updated model.User
	err := result.Decode(&updated)
	if err != nil {
		return nil, ResponseUnmarshallError{err: err}
	}

	return &updated, nil
}

// DeleteUser deletes the user with given id. If DB operation fails the unchanged error is returned.
func (m MongoUsersStorage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	var dbCtx, cancel = context.WithTimeout(ctx, m.dbTimeout)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$eq": id}}
	result, err := m.users.DeleteOne(dbCtx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return NotFoundError
	}

	return nil
}

func createGetUsersFilter(params model.GetUsersParams) bson.M {
	filter := bson.M{}
	if params.FilterFields.FirstName != "" {
		filter["first_name"] = params.FilterFields.FirstName
	}
	if params.FilterFields.LastName != "" {
		filter["last_name"] = params.FilterFields.LastName
	}
	if params.FilterFields.Nickname != "" {
		filter["nickname"] = params.FilterFields.Nickname
	}
	if params.FilterFields.Email != "" {
		filter["email"] = params.FilterFields.Email
	}
	if params.FilterFields.Country != "" {
		filter["country"] = params.FilterFields.Country
	}
	return filter
}

func createGetUsersOpts(params model.GetUsersParams) (*options.FindOptions, error) {
	if params.Sort.Field == "" {
		return nil, errors.New("sort field is required")
	}
	if params.PageSize < 0 {
		return nil, errors.New("page size cannot be negative number")
	}
	if params.Page < 0 {
		return nil, errors.New("page cannot be negative number")
	}

	//1 = ascending, -1 = descending
	sortType := 1
	if params.Sort.Type == "desc" {
		sortType = -1
	}
	sort := bson.D{{params.Sort.Field, sortType}}

	return options.Find().
		SetSort(sort).
		SetLimit(int64(params.PageSize)).
		SetSkip(int64(params.Page * params.PageSize)), nil
}
