package models

import (
	"errors"
)

// ErrNotFound is the internal representation used by the adapters to express a item was not found. (Instead of returning an empty <type>)
var ErrNotFound error

//ErrAlreadyExist is the error code when there is an attempt to register an already existing item in the database (User : name, IP : (value & project), Port : (ip & project)...)
var ErrAlreadyExist error

func init() {
	ErrNotFound = errors.New("Not found")
	ErrAlreadyExist = errors.New("Already exist")
}
