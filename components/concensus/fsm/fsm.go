package fsm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

// TODO: Optimise and check proper concurency
type ConfigFSM struct {
	mu  sync.RWMutex
	nc  *NodeConfig
	log *zap.Logger
}

// compile time interface conpatibility check
var _ raft.FSM = &ConfigFSM{}

func NewConfigFSM(log *zap.Logger) *ConfigFSM {
	return &ConfigFSM{
		log: log,
		nc:  &NodeConfig{},
	}
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (c *ConfigFSM) Apply(rlog *raft.Log) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Info("Apply", zap.Any("rLog", string(rlog.Data)))

	switch rlog.Type {
	case raft.LogCommand:
		var nc NodeConfig
		if err := json.Unmarshal(rlog.Data, &nc); err != nil {
			c.log.Error("Unable to marshal rlog.Data", zap.Error(err))
		}

		c.nc = &nc

		return c.nc
	}

	return nil
}

// Snapshot will be called during make snapshot.
// Snapshot is used to support log compaction.
func (c *ConfigFSM) Snapshot() (raft.FSMSnapshot, error) {
	// Get bytes of node config
	by, err := json.Marshal(c.nc)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(by), nil
}

// Restore is used to restore an FSM from a Snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (c *ConfigFSM) Restore(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	var nc NodeConfig
	err = json.Unmarshal(b, &nc)
	if err != nil {
		return err
	}

	c.nc = &nc
	return nil
}
