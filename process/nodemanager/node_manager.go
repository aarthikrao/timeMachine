package nodemanager

import (
	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
)

type NodeManager struct {
	selfNodeID string

	dsmgr  *dsm.DataStoreManager
	dhtMgr dht.DHT
}

func CreateNodeManager(dsmgr *dsm.DataStoreManager, dhtMgr dht.DHT) *NodeManager {
	return &NodeManager{
		dsmgr:  dsmgr,
		dhtMgr: dhtMgr,
	}
}

// Returns the location interface of the key. If the node is present on the same node,
// it returns the db, orelse it returns the connection to the respective server
func (nm *NodeManager) GetLocation(key string) (js.JobStore, error) {
	nodeVsSlot, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	// Fetch the node number if it exists
	slotNumber, ok := nodeVsSlot[nm.selfNodeID]
	if ok {
		return nm.dsmgr.GetDataNode(slotNumber)
	}

	return nil, nil
	// TODO: Return the DataStore interface with GRPC connection
	// that can be used to fetch data from another node

}
