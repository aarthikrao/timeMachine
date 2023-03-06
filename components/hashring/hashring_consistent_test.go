package hashring

import (
	"fmt"
	"reflect"
	"testing"
)

var cr *consistentHashing

func init() {
	fmt.Println("Initialising consistent hashing")
	cr = NewConsistentHashing(64, 21, []string{"node1", "node2", "node3"})
}

func Test_consistentHashing_GetKeyLocations(t *testing.T) {
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Key-A",
			args:    args{key: []byte("key-A")},
			want:    []string{"node3", "node1", "node2"},
			wantErr: false,
		}, {
			name:    "Key-{D}",
			args:    args{key: []byte("Key-{D}")},
			want:    []string{"node1", "node2", "node3"},
			wantErr: false,
		}, {
			name:    "vbhasdgfyt__89__jhgasdf",
			args:    args{key: []byte("vbhasdgfyt__89__jhgasdf")},
			want:    []string{"node1", "node2", "node3"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cr.GetKeyLocations(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("consistentHashing.GetKeyLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetKeyLocations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consistentHashing_GetNewNodeDelta(t *testing.T) {
	type args struct {
		nodeID string
	}
	tests := []struct {
		name                string
		args                args
		want                map[string][]int
		wantRelocPercentage float64
	}{
		{
			name: "Add node 4",
			args: args{nodeID: "node4"},
			want: map[string][]int{
				"node3": {51},
				"node4": {5, 27, 38, 40, 43, 45, 46, 53, 56, 57, 59, 61, 62},
			},
			wantRelocPercentage: 21.875,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := cr.GetNewNodeDelta(tt.args.nodeID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetNewNodeDelta() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantRelocPercentage {
				t.Errorf("consistentHashing.GetNewNodeDelta() relocationPercentage = %v, want %v", got1, tt.wantRelocPercentage)
			}
		})
	}
}

func Test_consistentHashing_GetRemoveNodeDelta(t *testing.T) {
	type args struct {
		nodeID string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]int
		want1   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := cr.GetRemoveNodeDelta(tt.args.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("consistentHashing.GetRemoveNodeDelta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetRemoveNodeDelta() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("consistentHashing.GetRemoveNodeDelta() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
