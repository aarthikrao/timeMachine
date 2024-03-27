package cordinator

import "errors"

var (
	ErrInvalidDetails = errors.New("invalid details")

	ErrRouteNotFound = errors.New("route not found")
)
