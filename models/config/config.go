package config

type Config struct {
	// The count of slots per node. Refer DHT package
	SlotPerNodeCount int `json:"slot_per_node_count,omitempty"`
}
