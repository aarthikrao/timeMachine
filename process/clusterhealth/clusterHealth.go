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
				continue
			}

			n.UnreachableCount++
			if n.UnreachableCount >= ch.UnreachableThreshold && !n.MarkedUnreachable {
				// The node has been down for more than UnreachableThreshold. We have to reassign the master
				ch.log.Info("Reassigning master slots because node is unreachable",
					zap.Int("threshold", n.UnreachableCount),
					zap.Time("lastContact", n.LastContact),
				)
				sn := ch.dht.ReassignMasterSlots(ni)

				by, err := consensus.ConvertConfigSnapshot(sn)
				if err != nil {
					ch.log.Error("Error in converting config snapshot", zap.Error(err))
				}

				if err = ch.cp.Apply(by); err != nil {
					ch.log.Error("Error in applying config snapshot to cp", zap.Error(err), zap.String("msg", string(by)))
				}

				n.MarkedUnreachable = true
			}
		}
	}
}
