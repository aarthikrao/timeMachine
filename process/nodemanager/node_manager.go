package nodemanager

import (
	"errors"

	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
	"github.com/aarthikrao/timeMachine/utils/address"
	"go.uber.org/zap"
)

var (
	// node has not yet been initalised
	ErrNotYetInitalised = errors.New("not yet initialised")
)

type NodeManager struct {
	selfNodeID dht.NodeID

	dsmgr *dsm.DataStoreManager

	connMgr *connectionmanager.ConnectionManager

	dhtMgr dht.DHT

	cp consensus.Consensus

	log *zap.Logger
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
	cp consensus.Consensus,
	log *zap.Logger,
) *NodeManager {
	return &NodeManager{
		selfNodeID: dht.NodeID(selfNodeID),
		dsmgr:      dsmgr,
		dhtMgr:     dhtMgr,
		connMgr:    connMgr,
		cp:         cp,
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

// Returns the location interface of the key. If the node is present on the same node,
// it returns the db, or else it returns the connection to the respective server
func (nm *NodeManager) GetJobStoreInterface(key string, leaderPreferred bool) (js.JobStore, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	// We process all requests via leader node.
	leader, follower, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	if leader.NodeID == nm.selfNodeID {
		// Give the db object
		return nm.dsmgr.GetDataNode(leader.SlotID)
	}

	// Data not present in this node. Give the connection object to the owner node
	if leaderPreferred {
		// Give the connection to the node with leader
		return nm.connMgr.GetJobStore(leader.NodeID)
	} else {
		// Give the connection to the node with follower
		return nm.connMgr.GetJobStore(follower.NodeID)
	}
}
