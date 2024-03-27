package config

type Config struct {
	Shards   int `json:"shards,omitempty" bson:"shards,omitempty"`
	Replicas int `json:"replicas,omitempty" bson:"replicas,omitempty"`
}
