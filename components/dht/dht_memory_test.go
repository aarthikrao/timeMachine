package dht

import (
	"reflect"
	"sort"
	"testing"
)

var d DHT

func init() {
	d = Create()

	dht, err := Initialise(4, []string{"node1", "node2", "node3"})
	if err != nil {
		panic(err)
	}
	d.Load(dht.GetSlotVsNodes())
}

func Test_dht_GetLocation(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		wantLeader   *SlotAndNode
		wantFollower *SlotAndNode
		wantErr      bool
	}{
		{
			name: "node1",
			key:  "Key-A",
			wantLeader: &SlotAndNode{
				SlotID: 0,
				NodeID: "node1",
			},
			wantFollower: &SlotAndNode{
				SlotID: 6,
				NodeID: "node2",
			},
		},
		{
			name: "node-{havskf8hgfh23##$%}",
			key:  "node-{havskf8hgfh23##$%}",
			wantLeader: &SlotAndNode{
				SlotID: 5,
				NodeID: "node2",
			},
			wantFollower: &SlotAndNode{
				SlotID: 11,
				NodeID: "node3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLeader, gotFollower, err := d.GetLocation(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("dht.GetLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLeader, tt.wantLeader) {
				t.Errorf("dht.GetLocation() gotLeader = %v, want %v", gotLeader, tt.wantLeader)
			}
			if !reflect.DeepEqual(gotFollower, tt.wantFollower) {
				t.Errorf("dht.GetLocation() gotFollower = %v, want %v", gotFollower, tt.wantFollower)
			}
		})
	}
}

// TestSnapshot checks if the Snapshot method returns the correct snapshot of the node vs slot ids map.
func TestSnapshot(t *testing.T) {
	// Initialize the DHT
	d, err := Initialise(2, []string{"node1", "node2"})
	if err != nil {
		t.Fatalf("Failed to initialize DHT: %v", err)
	}

	// Predefine a set of slots and load them into the DHT
	want := map[SlotID]*SlotInfo{
		SlotID(1): {NodeID("node1"), Leader},
		SlotID(2): {NodeID("node2"), Follower},
	}
	err = d.Load(want)
	if err != nil {
		t.Fatalf("Failed to load slots into DHT: %v", err)
	}

	// Use Snapshot to capture the current state of the DHT
	got := d.Snapshot()

	// Compare the captured state to the predefined set of slots
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot() = %v, want %v", got, want)
	}
}

// TestGetSlotsForNode checks if the GetSlotsForNode method returns the correct slots for a given node.
func TestGetSlotsForNode(t *testing.T) {
	// Initialize DHT
	d, _ := Initialise(2, []string{"node1", "node2", "node3"})

	// Define your expected result
	want := []SlotID{0, 1} // "node1" should have slots 0 and 1
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })

	// Get the result from the function
	got := d.GetSlotsForNode(NodeID("node1"))
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

	// Compare the result with your expected result
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetSlotsForNode() = %v, want %v", got, want)
	}
}

func TestDHTInitAndLoad(t *testing.T) {
	// Initialize DHT
	d, _ := Initialise(2, []string{"node1", "node2", "node3"})

	// Save the initial state
	initialState := d.Snapshot()

	// Create a new map of slot configurations
	newSlots := map[SlotID]*SlotInfo{
		SlotID(1): {NodeID("node1"), Leader},
		SlotID(2): {NodeID("node2"), Follower},
		SlotID(3): {NodeID("node1"), Follower},
		SlotID(4): {NodeID("node3"), Leader},
	}

	// Load the new configuration
	err := d.Load(newSlots)
	if err != nil {
		t.Fatalf("Failed to load new configuration: %v", err)
	}

	// Check if new configuration is loaded correctly
	for slotID, slotInfo := range newSlots {
		if got := d.GetSlotVsNodes()[slotID]; !reflect.DeepEqual(got, slotInfo) {
			t.Errorf("slotID %v: got %v, want %v", slotID, got, slotInfo)
		}
	}

	// Load the initial configuration back
	err = d.Load(initialState)
	if err != nil {
		t.Fatalf("Failed to load initial configuration: %v", err)
	}

	// Check if initial configuration is loaded correctly
	for slotID, slotInfo := range initialState {
		if got := d.GetSlotVsNodes()[slotID]; !reflect.DeepEqual(got, slotInfo) {
			t.Errorf("slotID %v: got %v, want %v", slotID, got, slotInfo)
		}
	}
}

// Tests if the dnt initialised flag is working
func TestDht_IsInitialised(t *testing.T) {
	d = Create()

	if d.IsInitialised() == true {
		t.Errorf("the dht is just created, it cannot be initalised")
	}

	dht, err := Initialise(4, []string{"node1", "node2", "node3"})
	if err != nil {
		panic(err)
	}

	if dht.IsInitialised() == false {
		t.Errorf("the dht has just be initialised and yet it returning it has not been initalised")
	}
}
