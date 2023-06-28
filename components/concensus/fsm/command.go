package fsm

import (
	"encoding/json"
)

type OperationType int

const (
	// In this case, the Data will contain the JSON snapshot of DHT map.
	SlotVsNodeChange OperationType = 1

	// Data will contain the JSON snapshot of the DHT map.
	// As opposed to SlotVsNodeChange, this message means that the nodes are being initialised for the first time
	InitialiseNodes OperationType = 2
)

type Command struct {
	Operation OperationType

	// Contains the rest of the data
	Data json.RawMessage
}
