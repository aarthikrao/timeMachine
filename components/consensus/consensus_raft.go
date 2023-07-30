package consensus

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"go.uber.org/zap"
)

// TODO: Review all this
const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

var _ Consensus = &raftConsensus{}

type raftConsensus struct {
	raft *raft.Raft
}

func NewRaftconsensus(serverID string, port int, volumeDir string, fsmStore raft.FSM, log *zap.Logger, bootstrap bool) (*raftConsensus, error) {
	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(serverID)
	raftConf.SnapshotThreshold = 1024

	store, err := raftboltdb.NewBoltStore(filepath.Join(volumeDir, "raft.dataRepo"))
	if err != nil {
		return nil, err
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(volumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		return nil, err
	}

	var raftBinAddr = fmt.Sprintf("127.0.0.1:%d", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBinAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		return nil, err
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	if bootstrap {
		// always start single server as a leader
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(serverID),
					Address: transport.LocalAddr(),
					// Suffrage: raft.Voter,
				},
			},
		}

		if err := raftServer.BootstrapCluster(configuration).Error(); err != nil {
			log.Error("Bootstrap error", zap.Error(err))
		}
	}

	return &raftConsensus{
		raft: raftServer,
	}, nil
}

// Join is called to add a new node in the cluster.
// It returns an error if this node is not a leader
func (r *raftConsensus) Join(nodeID, raftAddress string) error {
	if r.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	return r.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(raftAddress), 0, 0).Error()
}

// Remove is called to remove a particular node from the cluster.
// It returns an error if this node is not a leader
func (r *raftConsensus) Remove(nodeID string) error {
	if r.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	return r.raft.RemoveServer(raft.ServerID(nodeID), 0, 0).Error()
}

// Stats returns the stats of raft on this node
func (r *raftConsensus) Stats() map[string]string {
	return r.raft.Stats()
}

// Returns true if the current node is leader
func (r *raftConsensus) IsLeader() bool {
	return r.raft.State() == raft.Leader
}

// Returns address of the leader
func (r *raftConsensus) GetLeaderAddress() string {
	return string(r.raft.Leader())
}

// Apply is used to apply a command to the FSM
func (r *raftConsensus) Apply(cmd []byte) error {
	applyResponse := r.raft.Apply(cmd, 500*time.Millisecond)
	return applyResponse.Error()
}

// GetConfigurations returns the list of servers in the cluster
func (r *raftConsensus) GetConfigurations() ([]raft.Server, error) {
	cf := r.raft.GetConfiguration()
	if err := cf.Error(); err != nil {
		return nil, err
	}

	return cf.Configuration().Servers, nil
}
