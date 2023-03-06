package dht

type dht struct {
	// total number of slots
	slots int

	// maintains the location of all slots slotid vs nodeid
	nodes map[int]string
}

// Creates a new distributed hash table from the inputs.
// Should be called only from bootstrap mode or while creating a new cluster
func CreateDHT(slots int, nodes []string) (DHT, error) {

	return &dht{
		slots: slots,
	}, nil
}

// Loads data from a already existing configuration. This mus be called at the
// starting of the application and must be taken only from the master.
func Load(nodeVsSlots map[string][]int) (DHT, error) {

	return nil, nil
}

// Snapshot returns the node vs slot ids map.
func (d *dht) Snapshot() (nodeVsSlots map[string][]int, err error) {

	return nil, nil
}

// Returns the location of the primary and relica slots and corresponding nodes
func (d *dht) GetLocation() (slots []SlotInfo) {

	return nil
}

// MoveSlot reassigns the slot to a particular node.
// Only called after confirmation from master
func (d *dht) MoveSlot(slot int, fromNode, toNode string) (err error) {

	return nil
}
