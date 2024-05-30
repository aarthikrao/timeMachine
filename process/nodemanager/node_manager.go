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

	// some other node handles the leader shard
	ErrNotShardLeader = errors.New("not shard leader")
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
func (nm *NodeManager) InitAppDHT(shards, replicas int) error {
	servers, err := nm.cp.GetConfigurations()
	if err != nil {
		return err
	}
	var nodes []string
	for _, server := range servers {
		serverID := string(server.ID)
		nodes = append(nodes, serverID)
	}

	sn, err := dht.InitialiseDHT(shards, nodes, replicas)
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

func (nm *NodeManager) IsInitialised() error {
	shards := nm.dhtMgr.GetAllShardsForNode(nm.selfNodeID)
	if len(shards) > 0 {
		return dht.ErrDHTAlreadyInitialised
	}

	return nil
}

// InitialiseNode will be called once the dht is initialised.
// It will help setting up the connections and initialise the datastores
func (nm *NodeManager) InitialiseNode() error {
	shards := nm.dhtMgr.GetAllShardsForNode(nm.selfNodeID)
	if len(shards) <= 0 {
		return dht.ErrDHTNotInitialised
	}

	if err := nm.dsmgr.InitialiseDataStores(shards); err != nil {
		return err
	}

	if err := nm.createConnections(); err != nil {
		return err
	}

	// In a seperate routine keep running a poller to fetch jobs for the next minute and schedule it
	go func() {
		for range time.Tick(1 * time.Minute) {
			if err := nm.executeJobs(); err != nil {
				nm.log.Error("Unable to execute jobs", zap.Error(err))
			}
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
		if err := nm.connMgr.Add(serverID, grpcAddress); err != nil {
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

func (nm *NodeManager) GetLocalShard(shardID dht.ShardID) (js.JobStore, error) {
	if nm.dsmgr == nil {
		return nil, ErrNotYetInitalised
	}

	return nm.dsmgr.GetDataNode(shardID)
}

func (nm *NodeManager) GetRemoteConnection(nodeID dht.NodeID) (js.JobStoreWithReplicator, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	return nm.connMgr.GetJobStore(nodeID)
}

// Fetches the jobs for the next minute and schedules it to the executor
func (nm *NodeManager) executeJobs() error {
	for _, shardID := range nm.dhtMgr.GetLeaderShardsForNode(nm.selfNodeID) {
		js, err := nm.dsmgr.GetDataNode(shardID)
		if err != nil {
			return err
		}

		nextMinute := timeutil.GetCurrentMinutes() + 1

		jobs, err := js.FetchJobForBucket(nextMinute)
		if err != nil {
			return err
		}
		nm.exe.SetNextMin(int64(nextMinute))
		nm.log.Debug("Fetched jobs", zap.Any("jobs", jobs))
		for _, j := range jobs {
			nm.exe.Run(*j)
		}

	}

	return nil
}
