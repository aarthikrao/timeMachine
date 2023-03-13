package dht

import "errors"

var (
	// ErrMismatchedSlotsInfo indicates that the slot numbers may be duplicate or mismatched
	ErrMismatchedSlotsInfo = errors.New("mismatched slot info")

	// ErrDuplicateSlots indicates that there exsists duplicate slots in the input
	ErrDuplicateSlots = errors.New("duplicate slots")
)

type SlotInfo struct {
	Slot int    `json:"slot,omitempty"`
	Node string `json:"node,omitempty"`
}

// DHT contains the location of a given key in a distributed data system.
type DHT interface {

	// Snapshot returns the node vs slot ids map.
	Snapshot() (slotVsNode map[int]string)

	// Returns the location of the primary and relica slots and corresponding nodes
	GetLocation(key string) (slots []SlotInfo)

	// UpdateSlot reassigns the slot to a particular node.
	// Only called after confirmation from master
	UpdateSlot(slot int, fromNode, toNode string) (err error)

	// Returns a possible slot to migrate.
	Propose() (slot int, fromNode, toNode string, err error)
}
