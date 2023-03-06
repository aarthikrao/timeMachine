package dht

type SlotInfo struct {
	Slot int    `json:"slot,omitempty"`
	Node string `json:"node,omitempty"`
}

// DHT contains the location of a given key in a distributed data system.
type DHT interface {

	// Snapshot returns the node vs slot ids map.
	Snapshot() (nodeVsSlots map[string][]int, err error)

	// Returns the location of the primary and relica slots and corresponding nodes
	GetLocation() (slots []SlotInfo)

	// MoveSlot reassigns the slot to a particular node.
	// Only called after confirmation from master
	MoveSlot(slot int, fromNode, toNode string) (err error)
}
