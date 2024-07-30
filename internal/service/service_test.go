package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"user-service/internal/model"
)

// Unit tests that cover the User Creation logic. In a real project I would cover
// also all the remaining handlers. The tests would look very similar, therefore not writing them
// as I believe the existing ones should be enough to showcase the way to write them.

func Test_CreateUser(t *testing.T) {
	tests := []struct {
		name                   string
		user                   model.User
		dbError                error
		eventsError            error
		wantError              bool
		wantDBCreationCalled   bool
		wantEventPublishCalled bool
	}{
		{
			name: "happy path",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
				Email:     "valid@gmail.com",
			},
			wantDBCreationCalled:   true,
			wantEventPublishCalled: true,
		},
		{
			name: "DB user creation fails",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
				Email:     "valid@gmail.com",
			},
			dbError:              errors.New("DB error"),
			wantError:            true,
			wantDBCreationCalled: true,
		},
		{
			name: "Event publish failed - still seems as success to function caller",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
				Email:     "valid@gmail.com",
			},
			eventsError:            errors.New("events error"),
			wantDBCreationCalled:   true,
			wantEventPublishCalled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageMock := new(StorageMock)
			eventsMock := new(EventsProducerMock)

			ctx := context.Background()
			svc := New(storageMock, eventsMock)

			if tt.wantDBCreationCalled {
				storageMock.On("CreateUser", ctx, mock.MatchedBy(userCreationMatchFunc(tt.user))).Return(tt.dbError)
			}
			if tt.wantEventPublishCalled {
				eventsMock.On("Produce", mock.MatchedBy(userCreationEventMatchFunc(tt.user))).Return(tt.eventsError)
			}

			got, err := svc.CreateUser(ctx, tt.user)

			assert.Equal(t, tt.wantError, err != nil)
			if !tt.wantError {
				assert.True(t, userCreationMatchFunc(tt.user)(*got))
			}

			storageMock.AssertExpectations(t)
			eventsMock.AssertExpectations(t)
		})
	}
}

// userCreationMatchFunc matches user from CREATE request with the created one.
func userCreationMatchFunc(userToCreate model.User) func(gotUser model.User) bool {
	return func(gotUser model.User) bool {
		return gotUser.ID != uuid.UUID{} &&
			gotUser.FirstName == userToCreate.FirstName &&
			gotUser.LastName == userToCreate.LastName &&
			gotUser.Nickname == userToCreate.Nickname &&
			gotUser.Password == userToCreate.Password &&
			gotUser.Email == userToCreate.Email &&
			gotUser.Country == userToCreate.Country &&
			gotUser.CreatedAt.After(userToCreate.CreatedAt) &&
			gotUser.UpdatedAt.After(userToCreate.UpdatedAt)
	}
}

func userCreationEventMatchFunc(userToCreate model.User) func(event any) bool {
	return func(event any) bool {
		e, ok := event.(model.UserEvent)
		if !ok {
			return false
		}
		gotUser, ok := e.UserData.(model.User)
		if !ok {
			return false
		}

		return userCreationMatchFunc(userToCreate)(gotUser)
	}
}
