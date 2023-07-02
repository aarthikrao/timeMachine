package fsm

import "github.com/aarthikrao/timeMachine/components/dht"

// ConfigSnapshot is a snapshot of the current state of the node.
// It is replicated across all the nodes in the cluster with Raft.
type ConfigSnapshot struct {
	Slots map[dht.SlotID]*dht.SlotInfo
}
