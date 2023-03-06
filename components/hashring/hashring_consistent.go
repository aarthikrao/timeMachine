package hashring

import (
	"math/rand"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
)

type member string

func (m member) String() string {
	return string(m)
}

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type consistentHashing struct {
	c        *consistent.Consistent
	config   consistent.Config
	replicas int
}

func NewConsistentHashing(partitionCount int, replicationFactor int, existingNodes []string) *consistentHashing {
	var members []consistent.Member
	for _, existingNode := range existingNodes {
		members = append(members, member(existingNode))
	}

	cfg := consistent.Config{
		PartitionCount:    partitionCount,
		ReplicationFactor: replicationFactor,
		Load:              1.25,
		Hasher:            hasher{},
	}
	c := consistent.New(members, cfg)

	return &consistentHashing{
		c:        c,
		config:   cfg,
		replicas: 3,
	}
}

func (cr *consistentHashing) GetKeyLocations(key []byte) ([]string, error) {
	members, err := cr.c.GetClosestN(key, cr.replicas)
	if err != nil {
		return nil, err
	}

	var locations []string
	for _, member := range members {
		locations = append(locations, member.String())
	}

	return locations, nil
}

// GetNewNodeDelta creates a new instance of consistent hashring and returns the diff
// between current and new configuration so that partitions can be moved around
func (cr *consistentHashing) GetNewNodeDelta(nodeID string) (map[string][]int, float64) {
	oldMembers := cr.c.GetMembers()
	var newMembers []consistent.Member

	// Add old members
	newMembers = append(newMembers, oldMembers...)

	// Add new member
	m := member(nodeID)
	newMembers = append(newMembers, m)

	// Create a new ring
	c := consistent.New(newMembers, cr.config)

	// stores the delta paritions to move.
	delta := make(map[string][]int)

	// Change count
	var change float64

	for partID := 0; partID < cr.config.PartitionCount; partID++ {
		oldOwner := cr.c.GetPartitionOwner(partID).String()
		newOwner := c.GetPartitionOwner(partID).String()

		// if the partition has changed, add it to delta
		if oldOwner != newOwner {
			delta[newOwner] = append(delta[newOwner], partID)
			change++
		}
	}

	return delta, (change / float64(cr.config.PartitionCount) * 100)
}

func (cr *consistentHashing) AddNewNode(nodeID string) {
	// Add new member
	m := member(nodeID)
	cr.c.Add(m)
}

func (cr *consistentHashing) GetRemoveNodeDelta(nodeID string) (map[string][]int, int, error) {
	oldMembers := cr.c.GetMembers()
	var newMembers []consistent.Member

	for _, mem := range oldMembers {
		if mem.String() != nodeID {
			newMembers = append(newMembers, mem)
		}
	}

	if len(oldMembers) == len(newMembers) {
		// There has been no change
		return nil, 0, ErrNodeIDDoesntExist
	}

	// Create a new ring
	c := consistent.New(newMembers, cr.config)

	// stores the delta paritions to move.
	delta := make(map[string][]int)

	// Change count
	change := 0

	for partID := 0; partID < cr.config.PartitionCount; partID++ {
		oldOwner := cr.c.GetPartitionOwner(partID).String()
		newOwner := c.GetPartitionOwner(partID).String()

		// if the partition has changed, add it to delta
		if oldOwner != newOwner {
			delta[newOwner] = append(delta[newOwner], partID)
			change++
		}
	}

	return delta, (change / cr.config.PartitionCount * 100), nil
}

func (cr *consistentHashing) RemoveNode(nodeID string) {
	cr.c.Remove(nodeID)
}
