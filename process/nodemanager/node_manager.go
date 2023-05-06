package nodemanager

import (
	"errors"

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
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
) *NodeManager {
	return &NodeManager{
		selfNodeID: dht.NodeID(selfNodeID),
		dsmgr:      dsmgr,
		dhtMgr:     dhtMgr,
		connMgr:    connMgr,
	}
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
