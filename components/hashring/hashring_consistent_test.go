package hashring

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

var cr *consistentHashing

func init() {
	fmt.Println("Initialising consistent hashing")
	cr = NewConsistentHashing(271, 3, []string{"node1", "node2", "node3"})
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
			name:    "key1",
			args:    args{[]byte("key1")},
			want:    []string{"node3", "node1", "node2"},
			wantErr: false,
		}, {
			name:    "key2",
			args:    args{[]byte("key2")},
			want:    []string{"node2", "node3", "node1"},
			wantErr: false,
		}, {
			name:    "keyabc",
			args:    args{[]byte("keyabc")},
			want:    []string{"node2", "node3", "node1"},
			wantErr: false,
		}, {
			name:    "key#$%^&*#@*&%^",
			args:    args{[]byte("key#$%^&*#@*&%^")},
			want:    []string{"node1", "node2", "node3"},
			wantErr: false,
		}, {
			name:    "jhasd{bfu}aysf",
			args:    args{[]byte("jhasdbfuaysf")},
			want:    []string{"node2", "node3", "node1"},
			wantErr: false,
		}, {
			name:    "kjnsdiwuE {HFISAE}",
			args:    args{[]byte("key1")},
			want:    []string{"node3", "node1", "node2"},
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
		name string
		args args
		want map[string][]int
	}{
		{
			name: "Add 4th node",
			args: args{
				nodeID: "node4",
			},
			want: map[string][]int{
				"node1": {
					247,
				},
				"node2": {
					232, 233, 238, 239, 240, 242, 243, 245, 251, 252, 253, 254, 255,
					256, 257, 258, 259, 260, 261, 262, 263, 265, 267, 268, 269,
				},
				"node4": {
					1, 2, 6, 8, 9, 12, 15, 16, 17, 19, 23, 27, 28, 31, 32, 38, 39,
					42, 52, 55, 59, 60, 61, 72, 75, 77, 78, 79, 84, 87, 89, 90, 97, 98,
					99, 100, 103, 107, 110, 113, 120, 123, 124, 128, 130, 131, 132, 133,
					134, 135, 136, 139, 143, 144, 146, 150, 151, 153, 154, 156, 159, 161,
					163, 172, 175, 177, 178, 180, 182, 183, 184, 185, 186, 188, 197, 200,
					204, 207, 209, 212, 220, 225, 228, 230,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cr.GetNewNodeDelta(tt.args.nodeID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetNewNodeDelta() = %v, want %v", got, tt.want)
				by, _ := json.Marshal(got)
				fmt.Println(string(by))
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
		wantErr bool
	}{
		{
			name: "Remode 4th node",
			args: args{
				nodeID: "node4",
			},
			want: map[string][]int{
				"node1": {
					255, 256, 257, 258, 259, 260, 261, 262, 263, 265, 267, 268, 269,
				},
				"node2": {
					1, 6, 9, 12, 15, 17, 19, 23, 27, 28, 32, 38, 39, 42, 55, 59, 60,
					61, 72, 75, 77, 78, 79, 87, 89, 90, 97, 99, 100, 103, 110, 113,
					120, 123, 124, 128, 130, 131, 132, 134, 135, 136, 139, 143, 146,
					150, 151, 154, 156, 159, 163, 172, 175, 177, 178, 182, 183, 184,
					185, 186, 188, 197, 207, 209, 212, 225, 228, 230,
				},
				"node3": {
					2, 8, 16, 31, 52, 84, 98, 107, 133, 144, 153, 161, 180, 200, 204,
					220, 232, 233, 238, 239, 240, 242, 243, 245, 247, 251, 252, 253,
					254,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First add the noew node
			cr.AddNewNode(tt.args.nodeID)

			got, err := cr.GetRemoveNodeDelta(tt.args.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("consistentHashing.GetRemoveNodeDelta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetRemoveNodeDelta() = %v, want %v", got, tt.want)
			}
		})
	}
}
