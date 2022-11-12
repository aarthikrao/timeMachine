package datastore

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")

	ErrInvalidDataformat = errors.New("invalid data format")
)
