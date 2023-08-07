package wal

// WriteAheadLog logs all the changes in the disk. This way, all the writes are persisted even incase of failure
type WriteAheadLog interface {
	// Write writes the changes to the WAL.
	Write(change []byte) error

	// Replay calls the function f on all the records from the offset f
	Replay(offset int64, f func([]byte) error) error

	// GetLatestOffset returns the latest offset
	GetLatestOffset() (int64, error)

	// Close safely closes the WAL. All the data is persisted in WAL before closing
	Close() error
}
