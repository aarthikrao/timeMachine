package dht

import "errors"

var (
	// ErrMismatchedSlotsInfo indicates that the slot numbers may be duplicate or mismatched
	ErrMismatchedSlotsInfo = errors.New("mismatched slot info")

	// ErrDuplicateSlots indicates that there exsists duplicate slots in the input
	ErrDuplicateSlots = errors.New("duplicate slots")

	// ErrAlreadyInitialised indicates that you are trying to initialise a dht that has already been initialised
	ErrAlreadyInitialised = errors.New("dht already initialised")

	// ErrNotInitialised indicates that the dht is not initialised
	ErrNotInitialised = errors.New("dht not initialised")
)

// DHT contains the location of a given key in a distributed data system.
type DHT interface {

	// Creates a new distributed hash table from the inputs.
	// Should be called only from bootstrap mode or while creating a new cluster
	Initialise(slotCount int, nodes []string) error

	// Loads data from a already existing configuration.
	// This must be taken called after confirmation from the master
	Load(nodeVsSlots map[NodeID][]SlotID) error

	// Snapshot returns the node vs slot ids map.
	Snapshot() (slotVsNode map[SlotID]NodeID)

	// Returns the location of the primary and relica slots and corresponding nodes
	// map[SlotNumber]NodeID
	GetLocation(key string) (slots map[NodeID]SlotID, err error)

	// UpdateSlot reassigns the slot to a particular node.
	// Only called after confirmation from master
	UpdateSlot(slot SlotID, fromNode, toNode NodeID) (err error)

	// Returns a possible slot to migrate.
	Propose() (slot int, fromNode, toNode string, err error)
}
