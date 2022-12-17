package fsm

import (
	"fmt"

	"github.com/hashicorp/raft"
)

type snapshot struct {
	data []byte
}

func NewSnapshot(data []byte) *snapshot {
	return &snapshot{
		data: data,
	}
}

// Persist persist to disk. Return nil on success, otherwise return error.
func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write(s.data)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("sink.Write(): %v", err)
	}
	return sink.Close()
}

// Release release the lock after persist snapshot.
// Release is invoked when we are finished with the snapshot.
func (s *snapshot) Release() {
}
