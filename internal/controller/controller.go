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
	storage_err "user-service/internal/errors"
	"user-service/internal/model"
)

type Service interface {
	CreateUser(ctx context.Context, user model.User) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error)
	UpdateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// CreateUsersHandlers registers users endpoint paths with handlers to given router.
func CreateUsersHandlers(router *gin.RouterGroup, svc Service) {
	usersGroup := router.Group("users")
	usersGroup.POST("", createUser(svc))
	usersGroup.PUT(fmt.Sprintf(":%s", userIDPathParam), updateUser(svc))
	usersGroup.GET(fmt.Sprintf(":%s", userIDPathParam), getUser(svc))
	usersGroup.DELETE(fmt.Sprintf(":%s", userIDPathParam), deleteUser(svc))
	usersGroup.GET("", getUsers(svc))
}

// createUser returns a handler that handles user creation.
func createUser(svc Service) gin.HandlerFunc {
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

		createdUser, err := svc.CreateUser(c, user)
		if err != nil {
			logrus.WithError(err).
				WithField("user_id", user.ID).
				Error("failed to create user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
			c.Abort()
			return
		}

		c.JSON(http.StatusCreated, createdUser)
	}
}

// getUser returns a handler that handles user retrieval by ID.
func getUser(svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		user, err := svc.GetUserByID(c, userID)
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

// getUsers returns a handler that handles the users retrieval from the DB based on url params.
func getUsers(svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		params, err := parseGetUsersParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		users, err := svc.GetUsers(c, *params)
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

// updateUser returns a handler that handles user update.
func updateUser(svc Service) gin.HandlerFunc {
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

		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		user.ID = userID
		// db precision is in millis - doesn't support nanos
		user.UpdatedAt = time.Now().Truncate(time.Millisecond)

		err = svc.UpdateUser(c, user)
		if err != nil {
			if errors.Is(err, storage_err.NotFoundError) {
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

		c.Status(http.StatusNoContent)
	}
}

// deleteUser returns a handler that handles user removal.
func deleteUser(svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param(userIDPathParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("incorrect user ID format: %v", err.Error())})
			c.Abort()
			return
		}

		err = svc.DeleteUser(c, userID)
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
