package dht

import "errors"

var (
	// ErrMismatchedSlotsInfo indicates that the slot numbers may be duplicate or mismatched
	ErrMismatchedSlotsInfo = errors.New("mismatched slot info")

	// ErrDuplicateSlots indicates that there exists duplicate slots in the input
	ErrDuplicateSlots = errors.New("duplicate slots")

	// ErrAlreadyInitialised indicates that you are trying to initialise a dht that has already been initialised
	ErrAlreadyInitialised = errors.New("dht already initialised")

	// ErrNotInitialised indicates that the dht is not initialised
	ErrNotInitialised = errors.New("dht not initialised")
)

// DHT contains the location of a given key in a distributed data system.
type DHT interface {
	// GetLocation returns the location of the leader and follower slot and corresponding node
	GetLocation(key string) (leader *SlotAndNode, follower *SlotAndNode, err error)

	GetSlotsForNode(nodeID NodeID) []SlotID

	// Load loads data from an already existing configuration.
	// This must be taken called after confirmation from the master
	Load(slots map[SlotID]*SlotInfo) error

	// Snapshot returns the current node vs slot ids map
	Snapshot() map[SlotID]*SlotInfo

	IsInitialised() bool
}
