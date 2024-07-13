package storage

import (
	"errors"
	"fmt"
)

var NotFoundError = errors.New("not found")

// ResponseUnmarshallError defines state when DB write was successful but DB response unmarshal failed.
type ResponseUnmarshallError struct {
	err error
}

func (r ResponseUnmarshallError) Error() string {
	return fmt.Sprintf("failed to unmarshal data returned from DB: %s", r.err.Error())
}
