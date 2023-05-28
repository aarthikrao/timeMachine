package fsm

import "github.com/aarthikrao/timeMachine/components/dht"

type NodeConfig interface {
	// Returns the last updated time
	GetLastUpdatedTime() int

	GetNodeVsSlots() map[dht.NodeID][]dht.SlotID
}
