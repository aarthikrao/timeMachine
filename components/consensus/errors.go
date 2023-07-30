package consensus

import "errors"

var (
	ErrNotLeader = errors.New("not leader")
)
