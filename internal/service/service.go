package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
	custom_err "user-service/internal/errors"
	"user-service/internal/model"
)

type UsersStorage interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error)
	UpdateUser(ctx context.Context, user model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type EventsProducer interface {
	Produce(event any) error
}

type Service struct {
	storage        UsersStorage
	eventsProducer EventsProducer
}

func New(storage UsersStorage, eventsProducer EventsProducer) *Service {
	return &Service{
		storage:        storage,
		eventsProducer: eventsProducer,
	}
}

// CreateUser creates the User in DB and produces user created event.
func (s Service) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	newID, err := uuid.NewUUID()
	if err != nil {
		logrus.WithError(err).Error("failed to create UUID for new user")
		return nil, err
	}

	user.ID = newID
	// db precision is in millis - doesn't support nanos
	now := time.Now().Truncate(time.Millisecond)
	user.CreatedAt = now
	user.UpdatedAt = now

	if err = s.storage.CreateUser(ctx, user); err != nil {
		logrus.WithError(err).
			WithField("user_id", user.ID).
			Error("failed to create user")
		return nil, err
	}

	err = s.eventsProducer.Produce(model.NewUserCreatedEvent(user))
	if err != nil {
		// just log but return no error as this is just internal action that does not interest the caller of the func.
		logrus.WithError(err).
			WithField("user_id", user.ID).
			Error("failed to produce create user event")
	}

	return &user, nil
}

// GetUserByID retrieves the user from DB based on the provided id.
func (s Service) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.storage.GetUserByID(ctx, id)
	if err != nil {
		if !errors.Is(err, custom_err.NotFoundError) {
			logrus.WithError(err).
				WithField("user_id", id).
				Error("failed to get user")
		}

		return nil, err
	}

	return user, nil
}

// GetUsers retrieves the users from DB based on passed params.
func (s Service) GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error) {
	users, err := s.storage.GetUsers(ctx, params)
	if err != nil {
		logrus.WithError(err).Error("failed to get users")
		return nil, err
	}

	return users, nil
}

// UpdateUser updates the User in DB and produces user updated event.
func (s Service) UpdateUser(ctx context.Context, user model.User) error {
	// db precision is in millis - doesn't support nanos
	user.UpdatedAt = time.Now().Truncate(time.Millisecond)

	updated, err := s.storage.UpdateUser(ctx, user)
	if err != nil {
		var unmarshallErr custom_err.ResponseUnmarshallError
		if errors.As(err, &unmarshallErr) {
			// edge case - the User in the DB is updated but the DB response marshall failed.
			// Log the error but notify other systems about the change and don't fail as it was success from the caller POV.
			logrus.WithError(err).
				WithField("user_id", user.ID).
				Error("failed to unmarshall DB response")
		} else {
			logrus.WithError(err).
				WithField("user_id", user.ID).
				Error("failed to update user")
			return err
		}
	}

	err = s.eventsProducer.Produce(model.NewUserUpdatedEvent(*updated))
	if err != nil {
		// just log but return no error as this is just internal action that does not interest the caller of the func.
		logrus.WithError(err).
			WithField("user_id", user.ID.String()).
			Error("failed to produce update user event")
	}

	return nil
}

// DeleteUser deletes the User in DB and produces user deleted event.
func (s Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := s.storage.DeleteUser(ctx, id)
	if err != nil {
		logrus.WithError(err).
			WithField("user_id", id).
			Error("failed to delete user")
		return err
	}

	err = s.eventsProducer.Produce(model.NewUserDeletedEvent(id))
	if err != nil {
		// just log but return no error as this is just internal action that does not interest the caller of the func.
		logrus.WithError(err).
			WithField("user_id", id).
			Error("failed to produce delete user event")
	}

	return nil
}
