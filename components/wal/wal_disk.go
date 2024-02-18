package wal

import (
	"encoding/json"
	"time"

	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/wal"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

// logCommand specifies the type of operation for the wal command
type logCommand byte

var (
	SetLog    logCommand = 0x01
	DeleteLog logCommand = 0x02
)

// logEntry is the wal log entry
type logEntry struct {
	Data       []byte     `json:"data,omitempty"`
	Collection string     `json:"col,omitempty"`
	Operation  logCommand `json:"op,omitempty"`
}

type walMiddleware struct {
	w *wal.WriteAheadLog

	next jobstore.JobFetcher
}

// Compile time interface check
var _ jobstore.JobStoreOnDisk = (*walMiddleware)(nil)
var _ WALReader = (*walMiddleware)(nil)

// InitaliseWriteAheadLog returns a instance of WAL on disk
func InitaliseWriteAheadLog(
	walDir string,
	maxLogSize int64,
	maxSegments int,
	log *zap.Logger,
	next jobstore.JobFetcher,
) (*walMiddleware, error) {
	w, err := wal.NewWriteAheadLog(&wal.WALOptions{
		LogDir:            walDir,
		MaxLogSize:        maxLogSize,
		MaxSegments:       maxSegments,
		Log:               log,
		MaxWaitBeforeSync: 1 * time.Second, // TODO: change this variable to default
	})
	if err != nil {
		return nil, err
	}

	return &walMiddleware{
		w:    w,
		next: next, // the next interface to call
	}, nil
}

// Replay calls the function f on all the records from the offset f
func (d *walMiddleware) Replay(offset int64, f func([]byte) error) error {
	return d.w.Replay(offset, f)
}

// GetLatestOffset returns the latest offset
func (d *walMiddleware) GetLatestOffset() int64 {
	return d.w.GetOffset()
}

// Close safely closes the WAL. All the data is persisted in WAL before closing
func (d *walMiddleware) Close() error {
	d.w.Close()
	return d.next.Close()
}

func (d *walMiddleware) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	// This is just a dummy method. There is no implementation from wal side.
	// It only calls the fetch from next object
	return d.next.FetchJobForBucket(minute)
}

func (wm *walMiddleware) GetJob(collection, jobID string) (*jm.Job, error) {
	// During read operations, we dont write to wal
	return wm.next.GetJob(collection, jobID)
}

func (wm *walMiddleware) SetJob(collection string, job *jm.Job) error {
	by, err := job.ToBytes()
	if err != nil {
		errors.Wrap(err, "wal setjob job")
	}

	le := logEntry{
		Data:       by,
		Collection: collection,
		Operation:  SetLog,
	}

	if err := wm.makeEntry(le); err != nil {
		return err
	}

	return wm.next.SetJob(collection, job)
}

func (wm *walMiddleware) Type() jobstore.JobStoreType {
	return jobstore.WAL
}

func (wm *walMiddleware) DeleteJob(collection, jobID string) error {
	le := logEntry{
		Data:       []byte(jobID),
		Collection: collection,
		Operation:  DeleteLog,
	}

	if err := wm.makeEntry(le); err != nil {
		return err
	}

	return wm.next.DeleteJob(collection, jobID)
}

func (wm *walMiddleware) makeEntry(le logEntry) error {
	// TODO: Optimise to msgPack later
	entry, err := json.Marshal(le)
	if err != nil {
		errors.Wrap(err, "wal setjob entry")
	}

	_, err = wm.w.Write(entry)
	if err != nil {
		return err
	}

	return nil
}
