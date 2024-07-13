package model

import "github.com/google/uuid"

type Action string

const USER_CREATED Action = "created"
const USER_UPDATED Action = "updated"
const USER_DELETED Action = "deleted"

// UserEvent defines the event that is emitted by the service upon User data change.
type UserEvent struct {
	Action Action `json:"action"`
	// UserData is either User for create/update or UserDeletedData for delete events.
	UserData any `json:"user_data"`
}

type UserDeletedData struct {
	UserID uuid.UUID `json:"id"`
}

func NewUserCreatedEvent(userData User) UserEvent {
	return newUserEvent(USER_CREATED, userData)
}

func NewUserUpdatedEvent(userData User) UserEvent {
	return newUserEvent(USER_UPDATED, userData)
}

func NewUserDeletedEvent(userID uuid.UUID) UserEvent {
	return newUserEvent(USER_DELETED, UserDeletedData{UserID: userID})
}

func newUserEvent(action Action, userData any) UserEvent {
	return UserEvent{
		Action:   action,
		UserData: userData,
	}
}
