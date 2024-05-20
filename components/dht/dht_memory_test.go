package dht

import (
	"reflect"
	"testing"
)

var d *dht

func init() {
	d = Create() // Create an empty instance
	shards, err := InitialiseDHT(12, []string{"node1", "node2", "node3"}, 3)
	if err != nil {
		panic(err)
	}
	d.Load(shards) // load data to d
}

func Test_dht_GetShard(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    ShardLocation
		wantErr bool
	}{
		{
			name: "Key: ABCD",
			args: args{
				key: "ABCD",
			},
			want: ShardLocation{
				ID:        2,
				Leader:    "node3",
				Followers: []NodeID{"node1", "node2"},
			},
		}, {
			name: "Key: kg654fd89h",
			args: args{
				key: "kg654fd89h",
			},
			want: ShardLocation{
				ID:        5,
				Leader:    "node3",
				Followers: []NodeID{"node1", "node2"},
			},
		}, {
			name: "Key: )(*&^%$#@!aitgehv)",
			args: args{
				key: ")(*&^%$#@!aitgehv)",
			},
			want: ShardLocation{
				ID:        2,
				Leader:    "node3",
				Followers: []NodeID{"node1", "node2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.GetShard(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("dht.GetShard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dht.GetShard() = %v, want %v", got, tt.want)
			}
		})
	}
}
