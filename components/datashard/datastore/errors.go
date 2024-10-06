package datastore

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")

	ErrBucketNotFound = errors.New("bucket not found")

	ErrInvalidDataformat = errors.New("invalid data format")
)
