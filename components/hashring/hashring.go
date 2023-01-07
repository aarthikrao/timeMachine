// hashring uses consistent hashing to store the keys in a distributed sequence across the cluster denoted by the ring.
package hashring

// HashRing expects a consistent hash ring to distribute the keys across the cluster in a ring architecture.
type HashRing interface {

	// GetKeyLocations returns the location of key in the ring.
	GetKeyLocations(key []byte) ([]string, error)

	// GetNewNodeDelta returns the vnodes that have to be moved to new node.
	// This does not actually add the node to the cluster yet. The output will return the
	// existing nodeID vs list of partitions to be moved to the new nodeID
	// To add new node to the cluster, use AddNewNode method.
	GetNewNodeDelta(nodeID string) map[string][]int

	// AddNewNode adds a new node to the cluster. This method must be called only once
	// the data transfer to the new node has been confirmed. Any calls to GetKeyLocations post this call
	// will also include the newly added node
	AddNewNode(nodeID string)

	// GetRemoveNodeDelta returns the new node vs the list of partitions to be moved to them post removing
	// the node. This does not actually delete the node. Refer GetNewNodeDelta.
	// To remove the node from the cluster, use RemoveNode method.
	GetRemoveNodeDelta(nodeID string) (map[string][]int, error)

	// RemoveNode removes the node from the cluster. This method must be called only once
	// the data transfer to other nodes has been confirmed. Any calls to GetKeyLocations post this call
	// will not include the removed node
	RemoveNode(nodeID string)
}
