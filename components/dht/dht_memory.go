package dht

import (
	"github.com/cespare/xxhash/v2"
)

type SlotState string
type SlotID int
type NodeID string

var (
	Leader   SlotState = "leader"
	Follower SlotState = "follower"
)

type dht struct {
	// total number of slotCount
	slotCount int

	// maintains the location of all slots slotid vs nodeid
	slotVsNodes map[SlotID]NodeID
}

func Create() *dht {
	return &dht{}
}

// Creates a new distributed hash table from the inputs.
// Should be called only from bootstrap mode or while creating a new cluster
func (d *dht) Initialise(slotCount int, nodes []string) error {
	if len(d.slotVsNodes) > 0 {
		return ErrAlreadyInitialised
	}

	d.slotCount = slotCount
	d.slotVsNodes = make(map[SlotID]NodeID)

	nodeCount := len(nodes)
	distribution := make([]int, nodeCount)
	for i := 0; i < slotCount; i++ {
		distribution[i%nodeCount]++
	}

	slotNumber := 0
	for i := 0; i < len(distribution); i++ {
		for j := 0; j < distribution[i]; j++ {
			d.slotVsNodes[SlotID(slotNumber)] = NodeID(nodes[i])
			slotNumber++
		}
	}

	return nil
}

// Loads data from a already existing configuration.
// This must be taken called after confirmation from the master
func (d *dht) Load(nodeVsSlots map[NodeID][]SlotID) error {
	if len(d.slotVsNodes) > 0 {
		return ErrAlreadyInitialised
	}

	slotCount := 0
	d.slotVsNodes = make(map[SlotID]NodeID)

	for nodeID, slots := range nodeVsSlots {
		for _, slot := range slots {
			d.slotVsNodes[slot] = nodeID
			slotCount++
		}
	}

	if len(d.slotVsNodes) != slotCount {
		return ErrDuplicateSlots
	}
	d.slotCount = slotCount

	return nil
}

// Snapshot returns the node vs slot ids map.
func (d *dht) Snapshot() (slotVsNode map[SlotID]NodeID) {
	return d.slotVsNodes
}

// Returns the location of the primary and relica slots and corresponding nodes
func (d *dht) GetLocation(key string) (slots map[NodeID]SlotID, err error) {
	if len(d.slotVsNodes) == 0 {
		return nil, ErrNotInitialised
	}

	location1 := SlotID(d.hash(key) % d.slotCount)
	node1 := d.slotVsNodes[location1]

	// Finding the diagonally opposite replica
	location2 := SlotID((int(location1) + d.slotCount/2) % d.slotCount)
	node2 := d.slotVsNodes[location2]

	return map[NodeID]SlotID{
		node1: location1,
		node2: location2,
	}, nil
}

// UpdateSlot reassigns the slot to a particular node.
// Only called after confirmation from master
func (d *dht) UpdateSlot(slot SlotID, fromNode, toNode NodeID) (err error) {
	if len(d.slotVsNodes) == 0 {
		return ErrNotInitialised
	}

	// Confirm the current location
	if d.slotVsNodes[slot] != fromNode {
		return ErrMismatchedSlotsInfo
	}

	d.slotVsNodes[slot] = toNode
	return nil
}

// TODO
// Propose will choose a slot to move from a node which currently has the max number of slots.
func (d *dht) Propose() (slot int, fromNode, toNode string, err error) {
	if len(d.slotVsNodes) == 0 {
		return 0, "", "", ErrNotInitialised
	}

	return 0, "", "", nil
}

func (d *dht) hash(key string) int {
	return int(xxhash.Sum64([]byte(key)))
}
