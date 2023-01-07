package hashring

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_consistentHashing_GetKeyLocations(t *testing.T) {
	type args struct {
		key []byte
	}

	cr := NewConsistentHashing(271, 3, []string{"node1", "node2", "node3"})

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
			fmt.Println(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consistentHashing.GetKeyLocations() = %v, want %v", got, tt.want)
			}
		})
	}
}
