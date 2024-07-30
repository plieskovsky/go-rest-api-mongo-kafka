package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"user-service/internal/model"
)

type EventsProducerMock struct {
	mock.Mock
}

func (m *EventsProducerMock) Produce(event any) error {
	args := m.Called(event)
	return args.Error(0)
}

type StorageMock struct {
	mock.Mock
}

func (m *StorageMock) CreateUser(ctx context.Context, user model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *StorageMock) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *StorageMock) GetUsers(ctx context.Context, params model.GetUsersParams) ([]model.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *StorageMock) UpdateUser(ctx context.Context, user model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *StorageMock) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
