package clusterhealth

import (
	"time"

	"github.com/aarthikrao/timeMachine/components/consensus"
	dhtComponent "github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	"go.uber.org/zap"
)

type NodeHealth struct {
	LastContact time.Time

	// The number of times this node was unreachable
	UnreachableCount int

	// If the node has been marked unreachable, it means that we have already
	// reassigned the master-slots that belong to this node.
	// We are maintaining this variable to make sure we dont end up doing reassignment
	// multiple times
	MarkedUnreachable bool
}

type clusterHealth struct {
	dht     dhtComponent.DHT
	cp      consensus.Consensus
	connMgr *connectionmanager.ConnectionManager

	clusterInfo map[dhtComponent.NodeID]NodeHealth

	pollInterval         time.Duration
	UnreachableThreshold int

	log *zap.Logger
}

func CreateClusterHealthChecker(
	dht dhtComponent.DHT,
	cp consensus.Consensus,
	connMgr *connectionmanager.ConnectionManager,

	pollInterval time.Duration,
	UnreachableThreshold int,

	log *zap.Logger,
) *clusterHealth {
	ch := &clusterHealth{
		dht:                  dht,
		cp:                   cp,
		connMgr:              connMgr,
		pollInterval:         pollInterval,
		clusterInfo:          make(map[dhtComponent.NodeID]NodeHealth),
		UnreachableThreshold: UnreachableThreshold,
		log:                  log,
	}

	go ch.GetClusterHealth()

	return ch
}

// GetClusterHealth checks for health of the cluster only on the master node.
// It maintains a map of nodeID vs node health details. It reassigns master slots
// when nodes become unreachable beyond a threshold.
func (ch *clusterHealth) GetClusterHealth() {
	ticker := time.NewTicker(ch.pollInterval)

	for {
		<-ticker.C

		// Check if this node is the leader
		if !ch.cp.IsLeader() {
			continue
		}

		report := ch.connMgr.GetHealthStatus()
		ch.log.Info("Health check", zap.Any("report", report))

		for ni, reachable := range report {
			n := ch.clusterInfo[ni]

			if reachable {
				n.LastContact = time.Now()
				n.UnreachableCount = 0
				n.MarkedUnreachable = false

				ch.clusterInfo[ni] = n
				continue
			}

			ch.log.Info("Unreachable node",
				zap.String("unreachable", string(ni)),
				zap.Int("retryCount", n.UnreachableCount),
				zap.Bool("marked", n.MarkedUnreachable))

			n.UnreachableCount++
			if n.UnreachableCount >= ch.UnreachableThreshold && !n.MarkedUnreachable {
				ch.log.Info("Handling node failure",
					zap.Int("threshold", n.UnreachableCount),
					zap.Time("lastContact", n.LastContact),
					zap.String("unreachableNode", string(ni)),
				)

				ch.handleNodeFailure(ni)
				n.MarkedUnreachable = true
			}

			ch.clusterInfo[ni] = n
		}
	}
}

func (ch *clusterHealth) handleNodeFailure(ni dhtComponent.NodeID) {
	// Raft will take care of split brain

	snapshot := ch.dht.Snapshot()
	for si, loc := range snapshot {
		if loc.Leader == ni {

			// Here we are assiging the next follower as a leader
			// TODO: In real tife usecases, we will have to take into consideration
			// of a lot of factors before assigning the leader
			nextFollower := snapshot[si].Followers[0]
			shard := snapshot[si]
			shard.Leader = nextFollower

			snapshot[si] = shard
		}
	}

	by, err := consensus.ConvertConfigSnapshot(snapshot)
	if err != nil {
		ch.log.Error("Error in converting config snapshot", zap.Error(err))
	}

	if err = ch.cp.Apply(by); err != nil {
		ch.log.Error("Error in applying config snapshot to cp", zap.Error(err), zap.String("msg", string(by)))
	}
}
