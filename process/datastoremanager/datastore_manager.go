package datastoremanager

import (
	"errors"
	"fmt"
	"sync"

	ds "github.com/aarthikrao/timeMachine/components/datastore"
	"go.uber.org/zap"
)

var (
	ErrDataStoreNotInitialised = errors.New("data store not initialised")
)

type DataStoreManager struct {
	// This will implement JobStore interface(s)

	// This list will contain the nodes owned by this instance of the server
	slotsOwned map[int]ds.DataStoreConn

	// path to the parent directory containing all the data
	parentDirectory string

	mu  sync.RWMutex
	log *zap.Logger
}

// Creates the datastores for the nodes it owns.
func CreateDataStore(parentDirectory string, log *zap.Logger) *DataStoreManager {
	dsm := &DataStoreManager{
		parentDirectory: parentDirectory,
		log:             log,
	}

	return dsm
}

func (dsm *DataStoreManager) InitialiseDataStores(nodes []int) error {
	for _, i := range nodes {
		path := fmt.Sprintf("%s/%d.db", dsm.parentDirectory, i)
		datastore, err := ds.CreateBoltDataStore(path)
		if err != nil {
			return err
		}

		dsm.slotsOwned[i] = datastore
		dsm.log.Info("initialised node", zap.Int("nodeID", i), zap.String("path", path))
	}

	return nil
}

func (dsm *DataStoreManager) GetDataNode(nodeID int) (ds.DataStore, error) {
	dsm.mu.RLock()
	defer dsm.mu.RUnlock()

	if len(dsm.slotsOwned) <= 0 {
		return nil, ErrDataStoreNotInitialised
	}

	return dsm.slotsOwned[nodeID], nil
}

func (dsm *DataStoreManager) Close() {
	for _, db := range dsm.slotsOwned {
		db.Close()
	}
}
