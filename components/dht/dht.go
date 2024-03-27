package dht

import "errors"

var (
	ErrDHTNotInitialised     = errors.New("dht is not initialised")
	ErrReplicasLessThanNodes = errors.New("replicas are lesser than physical nodes")
)

type ShardID int
type NodeID string

type ShardLocation struct {
	ID        ShardID
	Leader    NodeID
	Followers []NodeID
}

// DHT contains the location of a given key in a distributed data system.
type DHT interface {
	// GetLocation returns the location of the leader and follower slot and corresponding node
	GetShard(key string) (ShardLocation, error)

	GetLeaderShardsForNode(nodeID NodeID) []ShardID

	GetAllShardsForNode(nodeID NodeID) []ShardID

	// Load loads data from an already existing configuration.
	// This must be taken called after confirmation from the master
	Load(shards map[ShardID]ShardLocation)

	// Snapshot returns the current node vs slot ids map
	Snapshot() map[ShardID]ShardLocation
}
