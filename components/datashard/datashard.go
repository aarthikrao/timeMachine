package datashard

import (
	"fmt"

	"github.com/aarthikrao/timeMachine/components/datashard/datastore"
	"github.com/aarthikrao/timeMachine/components/datashard/wal"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DataShard struct {
	wal   wal.WAL
	store jobstore.JobFetcher

	log *zap.Logger
}

var _ jobstore.JobFetcher = (*DataShard)(nil)

func InitialiseDataShard(slot dht.ShardID, parentDirectory string, log *zap.Logger) (datashard *DataShard, err error) {
	// Initialise the datastore
	path := fmt.Sprintf("%s/%d.db", parentDirectory, slot)
	ds, err := datastore.CreateBoltDataStore(path)
	if err != nil {
		return nil, err
	}

	// Initalise the wal and wrap it aroung the datastore
	walPath := fmt.Sprintf("%s/%d/", parentDirectory, slot)
	w, err := wal.InitaliseWriteAheadLog(
		walPath,
		10e6, // 10MB per file
		5,    // 5 files // TODOD: Move to config
		log,
	)
	if err != nil {
		return nil, err
	}

	log.Info("initialised data store node",
		zap.Int("slot", int(slot)),
		zap.String("path", path),
	)
	return &DataShard{
		wal:   w,
		store: ds,
		log:   log,
	}, nil
}

func (ds *DataShard) GetJob(collection, jobID string) (*jm.Job, error) {
	return ds.store.GetJob(collection, jobID)
}

func (ds *DataShard) SetJob(collection string, job *jm.Job) (offset int64, err error) {
	by, err := job.ToBytes()
	if err != nil {
		errors.Wrap(err, "wal set job")
	}

	le := wal.LogEntry{
		Operation:  wal.SetLog,
		Collection: collection,
		Data:       by,
	}

	offset, err = ds.wal.AddEntry(le)
	if err != nil {
		return 0, err
	}

	_, err = ds.store.SetJob(collection, job)
	if err != nil {
		return offset, err
	}

	return offset, nil
}

func (ds *DataShard) DeleteJob(collection, jobID string) (offset int64, err error) {
	le := wal.LogEntry{
		Operation:  wal.DeleteLog,
		Collection: collection,
		Data:       []byte(jobID),
	}

	offset, err = ds.wal.AddEntry(le)
	if err != nil {
		return 0, err
	}

	_, err = ds.store.DeleteJob(collection, jobID)
	if err != nil {
		return offset, err
	}

	return offset, nil
}

func (ds *DataShard) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	return ds.store.FetchJobForBucket(minute)
}

func (ds *DataShard) Close() error {
	if err := ds.wal.Close(); err != nil {
		return err
	}

	if err := ds.store.Close(); err != nil {
		return err
	}

	return nil
}
