package wal

import (
	"github.com/aarthikrao/wal"
	"go.uber.org/zap"
)

type diskWAL struct {
	w *wal.WriteAheadLog
}

// InitaliseWriteAheadLog returns a instance of WAL on disk
func InitaliseWriteAheadLog(
	walDir string,
	maxLogSize int64,
	maxSegments int,
	log *zap.Logger,
) (*diskWAL, error) {
	w, err := wal.NewWriteAheadLog(&wal.WALOptions{
		LogDir:      walDir,
		MaxLogSize:  maxLogSize,
		MaxSegments: maxSegments,
		Log:         log,
	})
	if err != nil {
		return nil, err
	}

	return &diskWAL{
		w: w,
	}, nil
}

// Write writes the changes to the WAL.
func (d *diskWAL) Write(change []byte) error {
	_, err := d.w.Write(change)
	return err
}

// Replay calls the function f on all the records from the offset f
func (d *diskWAL) Replay(offset int64, f func([]byte) error) error {
	return d.w.Replay(offset, f)
}

// GetLatestOffset returns the latest offset
func (d *diskWAL) GetLatestOffset() (int64, error) {
	return d.w.GetOffset()
}

// Close safely closes the WAL. All the data is persisted in WAL before closing
func (d *diskWAL) Close() error {
	return d.w.Close()
}
