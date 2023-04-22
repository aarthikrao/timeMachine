package fsm

import "github.com/aarthikrao/timeMachine/components/dht"

// NodeConfig stores all the configuration related to this node here.
// It is replicated across all the nodes in the cluster with Raft.
type nodeConfig struct {
	LastContactTime int `json:"last_contact_time,omitempty" bson:"last_contact_time,omitempty"`

	slotVsNode map[dht.NodeID][]dht.SlotID

	// NodeID vs address
	nodeAddress map[string]string
}
