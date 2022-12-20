package concensus

import "github.com/hashicorp/raft"

// Concensus is responsible for electing a leader, maintaining linearizability
// and maintaining the config and FSM in the cluster.
type Concensus interface {

	// Join is called to add a new node in the cluster.
	// It returns an error if this node is not a leader
	Join(nodeID, raftAddress string) error

	// Remove is called to remove a particular node from the cluster.
	// It returns an error if this node is not a leader
	Remove(nodeID string) error

	// Stats returns the stats of raft on this node
	Stats() map[string]string

	// Returns true if the current node is leader
	IsLeader() bool

	// Apply is used to apply a command to the FSM
	Apply(cmd []byte) error

	// GetConfigurations returns a list of servers that are currently in a cluster
	GetConfigurations() ([]raft.Server, error)
}
