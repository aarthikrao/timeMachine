package dht

import (
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

// dht is an instance that holds the information about the assignments of slots to nodes,
// along with their leadership roles.
type dht struct {

	// Maintains the location of all slots slotID vs nodeID.
	slotVsNodes map[SlotID]*SlotInfo

	mu sync.RWMutex
}

var _ DHT = &dht{}

// Create initializes an empty distributed hash table.
func Create() *dht {
	return &dht{}
}

// Initialise creates a new distributed hash table from the inputs.
// Should be called only from bootstrap mode or while creating a new cluster.
// For loading data, Initialise should always follow with Load().
func Initialise(slotCountperNode int, nodes []string) (map[SlotID]*SlotInfo, error) {
	// slotCountperNode is the number of slots assigned to each node,
	// and nodes should be a list of node IDs that participate in the dht.

	d := &dht{
		slotVsNodes: make(map[SlotID]*SlotInfo),
	}
	// nodeVsSlot contains the mapping of nodeID to slotID.
	nodeVsSlot := make(map[NodeID][]SlotID)

	nodeCount := len(nodes)
	slotCount := slotCountperNode * nodeCount

	// At the end of the distributing process,
	// distribution represents how many slots(the distribution count) are assigned to each node,
	// ensuring a relatively even distribution in a round-robin manner.
	distribution := make([]int, nodeCount)
	for i := 0; i < slotCount; i++ {
		distribution[i%nodeCount]++
	}

	// slotNumber will later be transformed to SlotID,
	// representing the ID of the distributed slot.
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

	// Assign leaders in a round-robin manner.
	// Each node in nodeVsSlot is assigned a leader and follower node for each of its allocated slots.
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

	return d.slotVsNodes, nil
}

// Load loads the slot info data from a already existing configuration to a given empty dht.
// This must be called only after confirmation from the master.
func (d *dht) Load(slots map[SlotID]*SlotInfo) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.slotVsNodes = slots // TODO: Verify if this will share memory
	return nil
}

// Snapshot returns a copy of d.slotVsNodes.
func (d *dht) Snapshot() map[SlotID]*SlotInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create a new map to hold the snapshot
	snapshot := make(map[SlotID]*SlotInfo)

	// Copy the data from d.slotVsNodes to the snapshot map
	for slotID, slotInfo := range d.slotVsNodes {
		snapshot[slotID] = slotInfo
	}

	return snapshot
}

func (d *dht) GetSlotsForNode(nodeID NodeID) []SlotID {
	d.mu.Lock()
	defer d.mu.Unlock()

	var slots []SlotID
	for slotID, slotInfo := range d.slotVsNodes {
		if slotInfo.NodeID == nodeID {
			slots = append(slots, slotID)
		}
	}

	return slots
}

// GetLocation returns the location of the leader and follower slots and their corresponding nodes
func (d *dht) GetLocation(key string) (leader *SlotAndNode, follower *SlotAndNode, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.slotVsNodes) == 0 {
		return nil, nil, ErrNotInitialised
	}

	slot1 := SlotID(d.hashSlot(key))
	node1 := d.slotVsNodes[slot1]
	sn1 := &SlotAndNode{
		SlotID: slot1,
		NodeID: node1.NodeID,
	}

	slot2 := d.replicaSlot(slot1)
	node2 := d.slotVsNodes[slot2]
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
