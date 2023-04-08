package connectionmanager

import (
	"errors"

	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/network"
)

var (
	ErrNodeAddressNotPresent = errors.New("node address not present")
)

type ConnectionManager struct {
	// This map will contain the collections to other nodes via the job store API
	connMgr map[string]js.JobStoreConn

	// contains the address of all the other time machine nodes.
	// This map has to be updated when a new node is added
	address map[string]string
}

func CreateConnectionManager() *ConnectionManager {
	// TODO: Initialise with seed node details
	return &ConnectionManager{}
}

// GetConnection returns an existing connection object.
// If the connection does not exist, it will create a new
// connection and return it.
func (cm *ConnectionManager) GetConnection(nodeID string) (js.JobStoreConn, error) {
	if conn, ok := cm.connMgr[nodeID]; ok {
		return conn, nil
	}

	// Connection is nnot present. Create a new connection
	nodeAddr, ok := cm.address[nodeID]
	if !ok {
		return nil, ErrNodeAddressNotPresent
	}

	return network.CreateConnection(nodeAddr)
}
