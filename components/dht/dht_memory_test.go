package dht

import (
	"reflect"
	"testing"
)

var d *dht

func init() {
	d = Create()
	slotsVsNodes, err := Initialise(4, []string{"node1", "node2", "node3"})
	d.Load(slotsVsNodes)
	if err != nil {
		panic(err)
	}
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
