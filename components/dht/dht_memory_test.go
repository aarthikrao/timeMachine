package dht

import (
	"fmt"
	"reflect"
	"testing"
)

var d *dht

func init() {
	var err error
	d, err = CreateDHT(12, []string{"node1", "node2", "node3"})
	if err != nil {
		panic(err)
	}
	fmt.Println("SlotCount", d.slotCount, " Nodes:", d.slotVsNodes)
}

func Test_dht_GetLocation(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantSlots []SlotInfo
	}{
		{
			name: "Key-A",
			args: args{key: "Key-A"},
			wantSlots: []SlotInfo{
				{
					Slot: 6,
					Node: "node2",
				}, {
					Slot: 0,
					Node: "node1",
				},
			},
		}, {
			name: "Key-{abcdefg}",
			args: args{key: "Key-{abcdefg}"},
			wantSlots: []SlotInfo{
				{
					Slot: 1,
					Node: "node1",
				}, {
					Slot: 7,
					Node: "node2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSlots := d.GetLocation(tt.args.key); !reflect.DeepEqual(gotSlots, tt.wantSlots) {
				t.Errorf("dht.GetLocation() = %v, want %v", gotSlots, tt.wantSlots)
			}
		})
	}
}
