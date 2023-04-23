package fsm

import "github.com/aarthikrao/timeMachine/components/dht"

type NodeConfig interface {
	// Returns the last updated time
	GetLastUpdatedTime() int

	GetNodeVsStruct() map[dht.NodeID][]dht.SlotID
}
