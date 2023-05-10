package connectionmanager

import (
	"sync"

	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/network"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	ErrNodeNotPresent = errors.New("node not present")
)

// This will aggregate all the connections and clients for
// the GRPC connection with other time machine node.
type timeMachineConnection struct {
	// The uri of the time machine instance
	address string

	// The main grpc connection that is created with another instance of time machine node
	grpcConn *grpc.ClientConn

	// All the clients
	jobStore js.JobStore
}

type ConnectionManager struct {

	// nodeID vs connection object
	tmcMap map[dht.NodeID]*timeMachineConnection
	mu     sync.RWMutex

	log *zap.Logger
}

// CreateConnectionManager returns the connection manager
// It does not initialise the connections. This will have to be done
// by using the AddNewConnection
func CreateConnectionManager(log *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		log:    log,
		tmcMap: make(map[dht.NodeID]*timeMachineConnection),
	}
}

// connects to the provided nodeID.
func (cm *ConnectionManager) connect(nodeID dht.NodeID, addr string) error {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock())

	if err != nil {
		return err
	}

	cm.tmcMap[nodeID] = &timeMachineConnection{
		address:  addr,
		grpcConn: conn,
		jobStore: network.CreateJobStoreClient(conn),
	}

	return nil
}

// Adds new connection to the connection manager
func (cm *ConnectionManager) AddNewConnection(nodeID string, address string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// TODO: Add retry mechanism
	return cm.connect(dht.NodeID(nodeID), address)
}

// GetJobStore returns an existing job store client
func (cm *ConnectionManager) GetJobStore(nodeID dht.NodeID) (js.JobStore, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if tmc, ok := cm.tmcMap[nodeID]; ok {
		return tmc.jobStore, nil
	}

	return nil, ErrNodeNotPresent
}

// Closes all the connections maintained by the connection manager
func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for nodeID, tmc := range cm.tmcMap {
		cm.log.Info("Closing connection with node",
			zap.String("nodeID", string(nodeID)),
			zap.String("addr", tmc.address),
		)

		tmc.grpcConn.Close()
	}
}
