package nodemanager

import (
	"errors"
	"time"

	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/executor"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
	"github.com/aarthikrao/timeMachine/utils/address"
	timeutil "github.com/aarthikrao/timeMachine/utils/time"
	"go.uber.org/zap"
)

var (
	// node has not yet been initalised
	ErrNotYetInitalised = errors.New("not yet initialised")

	// The jobstore corresponding to the key you are looking for does not exist on this node
	ErrNotSlotOwner = errors.New("not slot owner")

	// This means that this node is the owner of the slot and you are trying to access the remote connection to it
	ErrLocalSlotOwner = errors.New("local slot owner")
)

type NodeManager struct {
	selfNodeID dht.NodeID

	dsmgr *dsm.DataStoreManager

	connMgr *connectionmanager.ConnectionManager

	dhtMgr dht.DHT

	cp consensus.Consensus

	exe executor.Executor

	log *zap.Logger
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
	cp consensus.Consensus,
	exe executor.Executor,
	log *zap.Logger,
) *NodeManager {
	return &NodeManager{
		selfNodeID: dht.NodeID(selfNodeID),
		dsmgr:      dsmgr,
		dhtMgr:     dhtMgr,
		connMgr:    connMgr,
		cp:         cp,
		exe:        exe,
		log:        log,
	}
}

// Initialises the app DHT from the server list.
// It also publishes the slot and node map to other nodes via consensus module
func (nm *NodeManager) InitAppDHT(slotsPerNode int) error {
	servers, err := nm.cp.GetConfigurations()
	if err != nil {
		return err
	}
	var nodes []string
	for _, server := range servers {
		serverID := string(server.ID)
		nodes = append(nodes, serverID)
	}

	sn, err := dht.Initialise(slotsPerNode, nodes)
	if err != nil {
		nm.log.Error("Unable to initialise dht", zap.Error(err))
		return err
	}

	by, err := consensus.ConvertConfigSnapshot(sn)
	if err != nil {
		return err
	}

	return nm.cp.Apply(by)
}

// InitialiseNode will be called once the dht is initialised.
// It will help setting up the connections and initialise the datastores
func (nm *NodeManager) InitialiseNode() error {
	slots := nm.dhtMgr.GetSlotsForNode(nm.selfNodeID)
	if len(slots) <= 0 {
		return dht.ErrNotInitialised
	}

	if err := nm.dsmgr.InitialiseDataStores(slots); err != nil {
		return err
	}

	if err := nm.createConnections(); err != nil {
		return err
	}

	// In a seperate routine keep running a poller to fetch jobs for the next minute and schedule it
	go func() {
		for {
			if err := nm.executeJobs(); err != nil {
				nm.log.Error("Unable to execute jobs", zap.Error(err))
			}

			time.Sleep(1 * time.Minute)
		}
	}()

	nm.log.Info("Initialsed node")
	return nil
}

func (nm *NodeManager) createConnections() error {
	servers, err := nm.cp.GetConfigurations()
	if err != nil {
		return err
	}

	for _, server := range servers {
		serverID := string(server.ID)
		grpcAddress := address.GetGRPCAddress(string(server.Address))

		nm.log.Info("Connecting to GRPC server", zap.Any("id", server.ID), zap.String("address", grpcAddress))
		if err := nm.connMgr.AddNewConnection(serverID, grpcAddress); err != nil {
			nm.log.Error("Unable to add connection",
				zap.String("serverID", serverID),
				zap.String("address", grpcAddress),
				zap.Error(err),
			)
		} else {
			nm.log.Info("Added GRPC connection",
				zap.String("serverID", serverID),
				zap.String("addr", grpcAddress))
		}
	}
	return nil
}

func (nm *NodeManager) IsSlotOwner(key string, leaderRequired bool) (bool, error) {
	if nm.connMgr == nil {
		return false, ErrNotYetInitalised
	}

	leader, follower, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return false, err
	}

	if leaderRequired {
		if leader.NodeID == nm.selfNodeID {
			return true, nil
		}

	} else {
		if follower.NodeID == nm.selfNodeID {
			return true, nil
		}

	}

	return false, nil
}

func (nm *NodeManager) GetLocalSlot(key string, leaderRequired bool) (js.JobStore, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	leader, follower, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	if leaderRequired { // Return the leader datasource
		if leader.NodeID == nm.selfNodeID {
			// Give the db object
			return nm.dsmgr.GetDataNode(leader.SlotID)
		}

	} else { // Return the follower datasource
		if follower.NodeID == nm.selfNodeID {
			// Give the db object
			return nm.dsmgr.GetDataNode(follower.SlotID)
		}
	}

	return nil, ErrNotSlotOwner
}

func (nm *NodeManager) GetRemoteSlot(key string, leaderRequired bool) (js.JobStoreWithReplicator, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	leader, follower, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	if leaderRequired {
		// Check if the job exists locally
		if leader.NodeID == nm.selfNodeID {
			return nil, ErrLocalSlotOwner
		}

		// Give the connection to the node with leader
		return nm.connMgr.GetJobStore(leader.NodeID)

	} else {
		// Check if the job exists locally
		if follower.NodeID == nm.selfNodeID {
			return nil, ErrLocalSlotOwner
		}

		// Give the connection to the node with follower
		return nm.connMgr.GetJobStore(follower.NodeID)
	}
}

// Fetches the jobs fo the next minute and schedules it to the executor
func (nm *NodeManager) executeJobs() error {
	for _, slotID := range nm.dhtMgr.GetSlotsForNode(nm.selfNodeID) {
		js, err := nm.dsmgr.GetDataNode(slotID)
		if err != nil {
			return err
		}

		nextMinute := timeutil.GetCurrentMinutes() + 1

		jobs, err := js.FetchJobForBucket(nextMinute)
		if err != nil {
			return err
		}

		for _, j := range jobs {
			nm.exe.Run(*j)
		}

	}

	return nil
}
