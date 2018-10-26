package models

import (
	"errors"
)

// ErrNotFound is the internal representation used by the adapters to express a item was not found. (Instead of returning an empty <type>)
var ErrNotFound error

func init() {
	ErrNotFound = errors.New("Not found")
}
