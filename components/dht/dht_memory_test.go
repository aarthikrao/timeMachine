package dht

import (
	"reflect"
	"testing"
)

var d *dht

func init() {
	d = Create()
	err := d.Initialise(12, []string{"node1", "node2", "node3"})
	if err != nil {
		panic(err)
	}
}

func Test_dht_GetLocation(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantSlots []SlotInfo
		wantErr   bool
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
			wantErr: false,
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlots, err := d.GetLocation(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("dht.GetLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSlots, tt.wantSlots) {
				t.Errorf("dht.GetLocation() = %v, want %v", gotSlots, tt.wantSlots)
			}
		})
	}
}
