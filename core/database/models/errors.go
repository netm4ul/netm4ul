package models

import (
	"errors"
)

// ErrNotFound is the internal representation used by the adapters to express a item was not found. (Instead of returning an empty <type>)
var ErrNotFound error

//UserAlreadyExist is the error code when there is an attempt to register an already existing account
var ErrUserAlreadyExist error

func init() {
	ErrNotFound = errors.New("Not found")
	ErrUserAlreadyExist = errors.New("User already exist")
}
