// Implements a Distributed hash table
// See https://en.wikipedia.org/wiki/Distributed_hash_table

package dht

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cespare/xxhash/v2"
)

type SlotState string
type SlotID int
type NodeID string

var (
	Leader   SlotState = "leader"
	Follower SlotState = "follower"
)

type SlotInfo struct {
	NodeID    NodeID
	SlotState SlotState
}

type SlotAndNode struct {
	SlotID SlotID
	NodeID NodeID
}

// dht This struct has been set as an export function because it is a return type
// for exported function
type dht struct {

	// maintains the location of all slots slotid vs nodeid
	slotVsNodes map[SlotID]*SlotInfo

	mu sync.RWMutex
	// Initialization flag
	initialised bool
}

var _ DHT = &dht{}

// Create initializes an empty distributed hash table.
func Create() DHT {
	return &dht{}
}

// Initialise creates a new distributed hash table from the inputs.
// Should be called only from bootstrap mode or while creating a new cluster
func Initialise(slotCountperNode int, nodes []string) (DHT, error) {

	d := &dht{
		slotVsNodes: make(map[SlotID]*SlotInfo),
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// nodeVsSlot contains the mapping of nodeID to slotID.
	nodeVsSlot := make(map[NodeID][]SlotID)

	nodeCount := len(nodes)
	slotCount := slotCountperNode * nodeCount

	// The distribution below makes sure the slots are
	// assigned equally in a round robin manner
	distribution := make([]int, nodeCount)
	for i := 0; i < slotCount; i++ {
		distribution[i%nodeCount]++
	}

	slotNumber := 0
	for i := 0; i < len(distribution); i++ { // For each of the node
		for j := 0; j < distribution[i]; j++ { // For the distribution count assigned to that node
			nodeID := NodeID(nodes[i])
			slotID := SlotID(slotNumber)

			d.slotVsNodes[slotID] = &SlotInfo{
				NodeID: nodeID,
			}
			nodeVsSlot[nodeID] = append(nodeVsSlot[nodeID], slotID)
			slotNumber++
		}
	}

	// Assign leaders in a round robin manner
	for i := 0; i < slotCountperNode; i++ {
		for nodeID := range nodeVsSlot {
			slotID := nodeVsSlot[nodeID][i]
			slotInfo := d.slotVsNodes[slotID]
			if slotInfo.SlotState != "" {
				continue
			}
			slotInfo.SlotState = Leader

			replicaSlotInfo := d.slotVsNodes[d.replicaSlot(slotID)]
			replicaSlotInfo.SlotState = Follower
		}
	}
	d.initialised = true

	by, err := json.Marshal(d.slotVsNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal slot info: %w", err)
	}

	fmt.Printf("Slot Info: %s\n", string(by))

	return d, nil
}

// Load loads data from a already existing configuration.
// This must be called only after confirmation from the master
func (d *dht) Load(slots map[SlotID]*SlotInfo) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create a new map to avoid memory sharing issues
	d.slotVsNodes = make(map[SlotID]*SlotInfo)
	for k, v := range slots {
		d.slotVsNodes[k] = v
	}

	d.initialised = true

	return nil
}

// Snapshot returns the node vs slot ids map in json format
// A copy is returned here so the caller cannot modify the returned map
func (d *dht) Snapshot() map[SlotID]*SlotInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	slotVsNodesCopy := make(map[SlotID]*SlotInfo)

	for k, v := range d.slotVsNodes {
		slotVsNodesCopy[k] = &SlotInfo{
			NodeID:    v.NodeID,
			SlotState: v.SlotState,
		}
	}

	return slotVsNodesCopy
}

// GetSlotsForNode returns all slots for a specific node
func (d *dht) GetSlotsForNode(nodeID NodeID) []SlotID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var slots []SlotID
	for slotID, slotInfo := range d.slotVsNodes {
		if slotInfo.NodeID == nodeID {
			slots = append(slots, slotID)
		}
	}

	return slots
}

// IsInitialised returns if the hash table has been initialised
func (d *dht) IsInitialised() bool {
	return d.initialised
}

// GetLocation Returns the location of the leader and follower slots and their corresponding nodes
func (d *dht) GetLocation(key string) (leader *SlotAndNode, follower *SlotAndNode, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.initialised {
		return nil, nil, ErrNotInitialised
	}

	slot1 := SlotID(d.hashSlot(key))
	node1, ok := d.slotVsNodes[slot1]
	if !ok {
		return nil, nil, fmt.Errorf("slot %v not found", slot1)
	}
	sn1 := &SlotAndNode{
		SlotID: slot1,
		NodeID: node1.NodeID,
	}

	slot2 := d.replicaSlot(slot1)
	node2, ok := d.slotVsNodes[slot2]
	if !ok {
		return nil, nil, fmt.Errorf("slot %v not found", slot2)
	}
	sn2 := &SlotAndNode{
		SlotID: slot2,
		NodeID: node2.NodeID,
	}

	if node1.SlotState == Leader {
		return sn1, sn2, nil // sn1 is the leader
	} else {
		return sn2, sn1, nil // sn2 is the leader
	}
}

func (d *dht) hashSlot(key string) int {
	slotCount := uint64(len(d.slotVsNodes))
	hashValue := xxhash.Sum64([]byte(key))
	return int(hashValue % slotCount)
}

func (d *dht) replicaSlot(location1 SlotID) SlotID {
	slotCount := len(d.slotVsNodes)
	return SlotID((int(location1) + slotCount/2) % slotCount)
}
