package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/mail"
	"time"
	"user-service/internal/model"
	storage_err "user-service/internal/storage"
)

type UsersStorage interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error)
	UpdateUser(ctx context.Context, user model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type EventsProducer interface {
	Produce(event any) error
}

// CreateUsersHandlers registers users endpoint paths with handlers to given router.
func CreateUsersHandlers(router *gin.RouterGroup, storage UsersStorage, producer EventsProducer) {
	usersGroup := router.Group("users")
	usersGroup.POST("", createUser(storage, producer))
	usersGroup.PUT(fmt.Sprintf(":%s", userIDPathParam), updateUser(storage, producer))
	usersGroup.GET(fmt.Sprintf(":%s", userIDPathParam), getUser(storage))
	usersGroup.DELETE(fmt.Sprintf(":%s", userIDPathParam), deleteUser(storage, producer))
	usersGroup.GET("", getUsers(storage))
}

// createUser returns a handler that creates the user in the DB and produces new User creation event.
func createUser(storage UsersStorage, producer EventsProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if err := validateRequiredRequestFields(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		newID, err := uuid.NewUUID()
		if err != nil {
			logrus.WithError(err).Error("failed to create UUID for new user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
			c.Abort()
			return
		}

		user.ID = newID
		// db precision is in millis - doesn't support nanos
		now := time.Now().Truncate(time.Millisecond)
		user.CreatedAt = now
		user.UpdatedAt = now

		if err = storage.CreateUser(c, user); err != nil {
			logrus.WithError(err).
				WithField("user_id", user.ID).
				Error("failed to create user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
			c.Abort()
			return
		}

		err = producer.Produce(model.NewUserCreatedEvent(user))
		if err != nil {
			// just log but proceed with HTTP response as this is internal, non customer visible action
			logrus.WithError(err).
				WithField("user_id", user.ID).
				Error("failed to produce create user event")
		}

		c.JSON(http.StatusCreated, user)
	}
}

// getUser returns a handler that retrieves the user from the DB.
func getUser(storage UsersStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		user, err := storage.GetUser(c, userID)
		if err != nil {
			if errors.Is(err, storage_err.NotFoundError) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				c.Abort()
				return
			}
			logrus.WithError(err).
				WithField("user_id", userID).
				Error("failed to get user")
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, *user)
	}
}

// getUsers returns a handler that retrieves the users from the DB based on url params.
func getUsers(storage UsersStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		params, err := parseGetUsersParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		users, err := storage.GetUsers(c, *params)
		if err != nil {
			logrus.WithError(err).Error("failed to get users")
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		if len(users) == 0 {
			c.JSON(http.StatusOK, []model.User{})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

// updateUser returns a handler that updates the user in the DB and produces User updated event.
func updateUser(storage UsersStorage, producer EventsProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		produceEvent := true

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if err := validateRequiredRequestFields(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		user.ID = userID
		// db precision is in millis - doesn't support nanos
		user.UpdatedAt = time.Now().Truncate(time.Millisecond)

		updated, err := storage.UpdateUser(c, user)
		if err != nil {
			var unmarshallErr storage_err.ResponseUnmarshallError
			if errors.As(err, &unmarshallErr) {
				// edge case - the User in the DB is updated but the DB response marshall failed.
				// Log the error and skip event produce but create a success HTTP response because
				// this is a success from the caller point of view.
				produceEvent = false
				logrus.WithError(err).
					WithField("user_id", userID).
					Error("failed to unmarshall DB response")
			} else if errors.Is(err, storage_err.NotFoundError) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				c.Abort()
				return
			} else {
				logrus.WithError(err).
					WithField("user_id", userID).
					Error("failed to update user")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "user not updated"})
				c.Abort()
				return
			}
		}

		if produceEvent {
			err = producer.Produce(model.NewUserUpdatedEvent(*updated))
			if err != nil {
				// just log but proceed with HTTP response as this is internal, non customer visible action
				logrus.WithError(err).
					WithField("user_id", user.ID.String()).
					Error("failed to produce update user event")
			}
		}

		c.Status(http.StatusNoContent)
	}
}

// deleteUser returns a handler that removes the user from the DB and produces User deleted event.
func deleteUser(storage UsersStorage, producer EventsProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		err = storage.DeleteUser(c, userID)
		if err != nil {
			if errors.Is(err, storage_err.NotFoundError) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				c.Abort()
				return
			}
			logrus.WithError(err).
				WithField("user_id", userID).
				Error("failed to delete user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
			c.Abort()
			return
		}

		err = producer.Produce(model.NewUserDeletedEvent(userID))
		if err != nil {
			// just log but proceed with HTTP response as this is internal, non customer visible action
			logrus.WithError(err).
				WithField("user_id", userID).
				Error("failed to produce delete user event")
		}

		c.Status(http.StatusNoContent)
	}
}

func validateRequiredRequestFields(u model.User) error {
	if u.FirstName == "" {
		return errors.New("first name is required")
	}
	if u.LastName == "" {
		return errors.New("last name is required")
	}
	if u.Nickname == "" {
		return errors.New("nickname is required")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("email is invalid")
	}
	if u.Country == "" {
		return errors.New("country is required")
	}
	return nil
}
