package controller

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"user-service/internal/model"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *ServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *ServiceMock) GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *ServiceMock) UpdateUser(ctx context.Context, user model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *ServiceMock) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
