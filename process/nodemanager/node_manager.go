package nodemanager

import (
	"errors"

	"github.com/aarthikrao/timeMachine/components/concensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
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
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
	cp concensus.Concensus,
) *NodeManager {
	return &NodeManager{
		selfNodeID: dht.NodeID(selfNodeID),
		dsmgr:      dsmgr,
		dhtMgr:     dhtMgr,
		connMgr:    connMgr,
		cp:         cp,
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
