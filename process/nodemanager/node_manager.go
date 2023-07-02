package nodemanager

import (
	"errors"

	"github.com/aarthikrao/timeMachine/components/concensus"
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

	cp concensus.Concensus

	log *zap.Logger
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
	cp concensus.Concensus,
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

	return nm.dhtMgr.Initialise(slotsPerNode, nodes)
}

func (nm *NodeManager) CreateConnections() error {
	servers, err := nm.cp.GetConfigurations()
	if err != nil {
		return err
	}

	for _, server := range servers {
		serverID := string(server.ID)
		grpcAddress := address.GetGRPCAddress(string(server.Address))

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
// it returns the db, orelse it returns the connection to the respective server
func (nm *NodeManager) GetJobStoreInterface(key string) (js.JobStore, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	// We process all requests via leader node.
	leader, _, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	if leader.NodeID == nm.selfNodeID {
		// Give the db object
		return nm.dsmgr.GetDataNode(leader.SlotID)
	} else {
		// Give the connection to the node with leader
		return nm.connMgr.GetJobStore(leader.NodeID)
	}
}
