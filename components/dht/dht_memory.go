package dht

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

// TODO: We need to add permanent location map.
// It will tell us where the shards are located when there is no failover.
type dht struct {
	mu     sync.RWMutex
	shards map[ShardID]ShardLocation
}

var _ DHT = (*dht)(nil)

// Create initializes an empty distributed hash table.
func Create() *dht {
	return &dht{}
}

// Initialise creates a new distributed hash table from the inputs.
// Should be called only from bootstrap mode or while creating a new cluster.
func InitialiseDHT(shardCount int, seedNodes []string, replication int) (map[ShardID]ShardLocation, error) {
	shards := make(map[ShardID]ShardLocation, shardCount)
	nodeCount := len(seedNodes)

	if nodeCount < replication {
		return nil, ErrReplicasLessThanNodes
	}

	for i := 0; i < shardCount; i++ {
		leaderNode := NodeID(seedNodes[i%nodeCount])
		followerNodes := []NodeDetails{}

		for r := 1; r < replication; r++ {
			follower := NodeID(seedNodes[(i+r)%nodeCount])
			followerNodes = append(followerNodes, NodeDetails{
				ID: follower,
			})
		}

		shardID := ShardID(i)
		shards[shardID] = ShardLocation{
			ID: shardID,
			Leader: NodeDetails{
				ID: leaderNode,
			},
			Followers: followerNodes,
		}
	}

	return shards, nil
}

func (d *dht) GetShard(key string) (ShardLocation, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.shards[d.hashSlot(key)], nil
}

func (d *dht) hashSlot(key string) ShardID {
	slotCount := uint64(len(d.shards))
	hashValue := xxhash.Sum64([]byte(key))
	return ShardID(hashValue % slotCount)
}

func (d *dht) GetLeaderShardsForNode(nodeID NodeID) []ShardID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	leaderShards := []ShardID{}

	for shardID, shard := range d.shards {
		if shard.Leader.ID == nodeID {
			leaderShards = append(leaderShards, shardID)
		}
	}

	return leaderShards
}

func (d *dht) GetAllShardsForNode(nodeID NodeID) []ShardID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	shards := []ShardID{}

	for _, shardLocation := range d.shards {
		// Check for leader
		if shardLocation.Leader.ID == nodeID {
			shards = append(shards, shardLocation.ID)
		}

		// Check for follower shards
		for _, node := range shardLocation.Followers {
			if node.ID == nodeID {
				shards = append(shards, shardLocation.ID)
			}
		}
	}

	return shards
}

func (d *dht) Load(shards map[ShardID]ShardLocation) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.shards = shards
}

func (d *dht) Snapshot() map[ShardID]ShardLocation {
	d.mu.Lock()
	defer d.mu.Unlock()

	m := make(map[ShardID]ShardLocation)
	for shardID, shard := range d.shards {
		followers := []NodeDetails{}
		followers = append(followers, shard.Followers...)

		m[shardID] = ShardLocation{
			ID:        shard.ID,
			Leader:    shard.Leader,
			Followers: followers,
		}
	}

	return m
}
