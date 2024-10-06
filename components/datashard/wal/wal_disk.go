package wal

import (
	"encoding/json"
	"time"

	"github.com/aarthikrao/wal"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type walStore struct {
	w *wal.WriteAheadLog
}

var _ WAL = (*walStore)(nil)

// InitaliseWriteAheadLog returns a instance of WAL on disk
func InitaliseWriteAheadLog(
	walDir string,
	maxLogSize int64,
	maxSegments int,
	log *zap.Logger,
) (*walStore, error) {
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

	return &walStore{
		w: w,
	}, nil
}

func (wm *walStore) AddEntry(le LogEntry) (offset int64, err error) {
	entry, err := json.Marshal(le)
	if err != nil {
		errors.Wrap(err, "wal setjob entry")
	}

	return wm.w.Write(entry)
}

// Replay calls the function f on all the records from the offset f
func (ws *walStore) Replay(offset int64, f func([]byte) error) error {
	return ws.w.Replay(offset, f)
}

// GetLatestOffset returns the latest offset
func (ws *walStore) GetLatestOffset() int64 {
	return ws.w.GetOffset()
}

// Close safely closes the WAL. All the data is persisted in WAL before closing
func (ws *walStore) Close() error {
	return ws.w.Close()

}
