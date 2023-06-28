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
	nc *NodeConfigHolder

	lastUpdateTime int

	mu  sync.RWMutex
	log *zap.Logger
}

var (
	// compile time interface conpatibility check
	//
	// ConfigFSM should implement both the raft
	// interface and the NodeConfig interface.
	_ raft.FSM   = &ConfigFSM{}
	_ NodeConfig = &ConfigFSM{}
)

func NewConfigFSM(log *zap.Logger) *ConfigFSM {
	return &ConfigFSM{
		log: log,
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
		err := c.handleChange(rlog.Data)
		if err != nil {
			c.log.Error("error in applying log command", zap.Error(err))
		}
	}

	return nil
}

// Snapshot will be called during make snapshot.
// Snapshot is used to support log compaction.
func (c *ConfigFSM) Snapshot() (raft.FSMSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

	var nc NodeConfigHolder
	err = json.Unmarshal(b, &nc)
	if err != nil {
		return err
	}

	c.nc = &nc
	return nil
}

// handleChange calls the respective method handler when there is a change
func (c *ConfigFSM) handleChange(data []byte) error {

	var cmd Command
	err := json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	switch cmd.Operation {
	case SlotVsNodeChange:
		// Decode the data change for slot vs node change

	}

	return nil
}

func (c *ConfigFSM) GetLastUpdatedTime() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastUpdateTime
}
