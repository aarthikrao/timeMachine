package fsm

import (
	"encoding/json"
)

type OperationType int

const (
	SlotVsNodeChange OperationType = 1
)

type Command struct {
	Operation OperationType

	// Contains the rest of the data
	Data json.RawMessage
}
