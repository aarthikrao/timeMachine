package fsm

// NodeConfig stores all the configuration related to this node here.
// It is replicated across all the nodes in the cluster with Raft.
type NodeConfig struct {
	// NodeID vs IP
	NodeList map[string]string
}
