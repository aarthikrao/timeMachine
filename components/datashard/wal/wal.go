package wal

// LogEntry is the wal log entry
type LogEntry struct {
	Data       []byte     `json:"data,omitempty"`
	Collection string     `json:"col,omitempty"`
	Operation  LogCommand `json:"op,omitempty"`
}

// LogCommand specifies the type of operation for the wal command
type LogCommand byte

var (
	SetLog    LogCommand = 0x01
	DeleteLog LogCommand = 0x02
)

// WAL reads all the changes from the disk
type WAL interface {
	AddEntry(entry LogEntry) (int64, error)

	// Replay calls the function f on all the records from the offset f
	Replay(offset int64, f func([]byte) error) error

	// GetLatestOffset returns the latest offset
	GetLatestOffset() int64

	// Close safely closes the WAL. All the data is persisted in WAL before closing
	Close() error
}
