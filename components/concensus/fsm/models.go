package fsm

import (
	"encoding/json"

	"github.com/aarthikrao/timeMachine/components/dht"
)

type OperationType int

const (
	// In this case, the Data will contain the JSON snapshot of DHT map.
	SlotVsNodeChange OperationType = 1

	// Data will contain the JSON snapshot of the DHT map.
	// As opposed to SlotVsNodeChange, this message means that the nodes are being initialised for the first time
	InitialiseNodes OperationType = 2
)

// This is a wrapper to propagate the changes to all nodes
type Command struct {
	Operation OperationType `json:"operation,omitempty" bson:"operation,omitempty"`

	// Contains the rest of the data
	Data json.RawMessage `json:"data,omitempty" bson:"data,omitempty"`
}

// ConfigSnapshot is a snapshot of the current state of the node.
// It is replicated across all the nodes in the cluster with Raft.
type ConfigSnapshot struct {
	Slots map[dht.SlotID]*dht.SlotInfo `json:"slots,omitempty" bson:"slots,omitempty"`
}
