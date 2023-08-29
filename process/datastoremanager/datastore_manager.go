package datastoremanager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aarthikrao/timeMachine/components/datastore"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/wal"
	"go.uber.org/zap"
)

var (
	ErrDataStoreNotInitialised = errors.New("data store not initialised")
)

type DataStoreManager struct {
	// This will implement JobStore interface(s)

	// This list will contain the nodes owned by this instance of the server
	slotsOwned map[dht.SlotID]js.JobStoreConn

	// path to the parent directory containing all the data
	parentDirectory string

	mu  sync.RWMutex
	log *zap.Logger
}

// Creates the datastores for the nodes it owns.
func CreateDataStore(parentDirectory string, log *zap.Logger) *DataStoreManager {
	dsm := &DataStoreManager{
		parentDirectory: parentDirectory,
		slotsOwned:      make(map[dht.SlotID]js.JobStoreConn),
		log:             log,
	}

	return dsm
}

func (dsm *DataStoreManager) InitialiseDataStores(slots []dht.SlotID) error {
	for _, slot := range slots {
		ds, err := dsm.initialiseVNode(slot)
		if err != nil {
			return err
		}

		dsm.slotsOwned[slot] = ds
	}

	return nil
}

func (dsm *DataStoreManager) initialiseVNode(slot dht.SlotID) (js jobstore.JobStoreConn, err error) {
	// Initialise the datastore
	path := fmt.Sprintf("%s/%d.db", dsm.parentDirectory, slot)
	ds, err := datastore.CreateBoltDataStore(path)
	if err != nil {
		return nil, err
	}

	// Initalise the wal and wrap it aroung the datastore
	walPath := fmt.Sprintf("%s/%d/", dsm.parentDirectory, slot) // TODO: finalise
	js, err = wal.InitaliseWriteAheadLog(
		walPath,
		10e6, // 10MB per file
		5,    // 5 files // TODOD: Move to config
		dsm.log,
		ds,
	)
	if err != nil {
		return nil, err
	}

	dsm.log.Info("initialised data store node",
		zap.Int("slot", int(slot)),
		zap.String("path", path),
	)
	return js, nil
}

func (dsm *DataStoreManager) GetDataNode(slotID dht.SlotID) (js.JobStore, error) {
	dsm.mu.RLock()
	defer dsm.mu.RUnlock()

	if len(dsm.slotsOwned) <= 0 {
		return nil, ErrDataStoreNotInitialised
	}

	return dsm.slotsOwned[slotID], nil
}

func (dsm *DataStoreManager) Close() {
	for _, db := range dsm.slotsOwned {
		db.Close()
	}
}
