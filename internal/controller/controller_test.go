package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/internal/model"
)

// Unit tests that cover the User Creation handler logic. In a real project I would cover
// also all the remaining handlers. The tests would look very similar, therefore not writing them
// as I believe the existing ones should be enough to showcase the way to write them.

func Test_CreateUserHandler(t *testing.T) {
	tests := []struct {
		name              string
		payload           model.User
		stringPayload     string
		serviceError      error
		wantStatusCode    int
		wantFailureBody   string
		wantServiceCalled bool
	}{
		{
			name: "happy path",
			payload: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
				Email:     "valid@gmail.com",
			},
			wantStatusCode:    http.StatusCreated,
			wantServiceCalled: true,
		},
		{
			name: "invalid payload - missing firstname",
			payload: model.User{
				LastName: "valid",
				Nickname: "valid",
				Password: "valid",
				Country:  "valid",
				Email:    "valid@gmail.com",
			},
			wantStatusCode:  http.StatusBadRequest,
			wantFailureBody: "{\"error\":\"first name is required\"}",
		},
		{
			name: "Service call fails",
			payload: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
				Email:     "valid@gmail.com",
			},
			serviceError:      errors.New("DB error"),
			wantStatusCode:    http.StatusInternalServerError,
			wantFailureBody:   "{\"error\":\"user not created\"}",
			wantServiceCalled: true,
		},
		{
			name:              "invalid body",
			stringPayload:     "invalid payload",
			wantStatusCode:    http.StatusBadRequest,
			wantServiceCalled: false,
			wantFailureBody:   "{\"error\":\"invalid character 'i' looking for beginning of value\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := new(ServiceMock)

			createUserHandler := createUser(serviceMock)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			requestPayload, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			var reqPayload io.Reader
			if tt.stringPayload != "" {
				reqPayload = bytes.NewBuffer([]byte(tt.stringPayload))
			} else {
				reqPayload = bytes.NewReader(requestPayload)
			}

			ctx.Request = &http.Request{Body: io.NopCloser(reqPayload)}

			if tt.wantServiceCalled {
				serviceMock.On("CreateUser", ctx, tt.payload).Return(&tt.payload, tt.serviceError)
			}

			// call the handler
			createUserHandler(ctx)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantStatusCode == http.StatusCreated {
				var createdUser model.User
				err := json.Unmarshal(w.Body.Bytes(), &createdUser)
				require.NoError(t, err)
				require.Equal(t, tt.payload, createdUser)
			} else {
				assert.Equal(t, tt.wantFailureBody, w.Body.String())
			}

			serviceMock.AssertExpectations(t)
		})
	}
}

func Test_validateRequiredRequestFields(t *testing.T) {
	tests := []struct {
		name          string
		user          model.User
		wantErr       bool
		wantErrString string
	}{
		{
			name: "valid user",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Email:     "valid@gmail.com",
				Country:   "valid",
			},
			wantErr: false,
		},
		{
			name: "firstname missing user",
			user: model.User{
				FirstName: "",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Email:     "valid@gmail.com",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "first name is required",
		},
		{
			name: "last name missing",
			user: model.User{
				FirstName: "valid",
				Nickname:  "valid",
				Password:  "valid",
				Email:     "valid@gmail.com",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "last name is required",
		},
		{
			name: "nickname missing",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Password:  "valid",
				Email:     "valid@gmail.com",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "nickname is required",
		},
		{
			name: "password missing",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Email:     "valid@gmail.com",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "password is required",
		},
		{
			name: "email missing",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "email is required",
		},
		{
			name: "email invalid",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Email:     "invalid",
				Country:   "valid",
			},
			wantErr:       true,
			wantErrString: "email is invalid",
		},
		{
			name: "country missing",
			user: model.User{
				FirstName: "valid",
				LastName:  "valid",
				Nickname:  "valid",
				Password:  "valid",
				Email:     "valid@gmail.com",
			},
			wantErr:       true,
			wantErrString: "country is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateRequiredRequestFields(tt.user)

			assert.Equal(t, tt.wantErr, gotErr != nil)
			if tt.wantErr {
				assert.Equal(t, gotErr.Error(), tt.wantErrString)
			}
		})
	}
}
