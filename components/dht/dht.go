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

	// Returns the location of the primary and relica slots and corresponding nodes
	// map[SlotNumber]NodeID
	GetLocation(key string) (slots map[NodeID]SlotID, err error)

	// Load Loads data from a already existing configuration.
	// This must be taken called after confirmation from the master
	// Snapshot returns the current node vs slot ids map
	// Both the methods use json format.
	Load(data []byte) error
	Snapshot() (data []byte, err error)
}
