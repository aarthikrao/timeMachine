package datastoremanager

import (
	"errors"
	"sync"

	"github.com/aarthikrao/timeMachine/components/datashard"
	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"go.uber.org/zap"
)

var (
	ErrDataStoreNotInitialised = errors.New("data store not initialised")
)

type DataStoreManager struct {
	// This list will contain the nodes owned by this instance of the server
	slotsOwned map[dht.ShardID]js.JobFetcher

	// path to the parent directory containing all the data
	parentDirectory string

	mu  sync.RWMutex
	log *zap.Logger
}

// Creates the datastores for the nodes it owns.
func CreateDataStore(parentDirectory string, log *zap.Logger) *DataStoreManager {
	dsm := &DataStoreManager{
		parentDirectory: parentDirectory,
		slotsOwned:      make(map[dht.ShardID]js.JobFetcher),
		log:             log,
	}

	return dsm
}

func (dsm *DataStoreManager) InitialiseDataStores(slots []dht.ShardID) error {
	for _, slot := range slots {
		ds, err := datashard.InitialiseDataShard(slot, dsm.parentDirectory, dsm.log)
		if err != nil {
			return err
		}

		dsm.slotsOwned[slot] = ds
	}

	return nil
}

func (dsm *DataStoreManager) GetDataNode(slotID dht.ShardID) (js.JobFetcher, error) {
	dsm.mu.RLock()
	defer dsm.mu.RUnlock()

	if len(dsm.slotsOwned) <= 0 {
		return nil, ErrDataStoreNotInitialised
	}

	return dsm.slotsOwned[slotID], nil // TODO: Handle node not available
}

func (dsm *DataStoreManager) Close() {
	for _, db := range dsm.slotsOwned {
		db.Close()
	}
}
