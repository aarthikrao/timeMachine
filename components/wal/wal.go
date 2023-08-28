package wal

// WALReader reads all the changes from the disk
type WALReader interface {

	// Replay calls the function f on all the records from the offset f
	Replay(offset int64, f func([]byte) error) error

	// GetLatestOffset returns the latest offset
	GetLatestOffset() int64

	// Close safely closes the WAL. All the data is persisted in WAL before closing
	Close() error
}
