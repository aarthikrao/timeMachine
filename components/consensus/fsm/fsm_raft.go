package fsm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/routestore"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

// TODO: Optimise and check proper concurency
type ConfigFSM struct {
	lastUpdateTime int

	dht    dht.DHT
	rStore *routestore.RouteStore

	// This function will be called by the config FSM when a change in configuration occurs.
	// You can use this function to update the node connections etc.
	onChangeHandler func() error

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

func NewConfigFSM(
	dht dht.DHT,
	rStore *routestore.RouteStore,
	log *zap.Logger,
) *ConfigFSM {
	return &ConfigFSM{
		dht:    dht,
		rStore: rStore,
		log:    log,
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
	// TODO: Add snapshot of current node configuration

	return NewSnapshot(nil), nil // Add bytes here
}

// Restore is used to restore an FSM from a Snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (c *ConfigFSM) Restore(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	var cs ConfigSnapshot
	err = json.Unmarshal(b, &cs)
	if err != nil {
		return err
	}

	c.handleSlotNodeChange(&cs)
	return nil
}

// handleChange calls the respective method handler when there is a change
func (c *ConfigFSM) handleChange(data []byte) error {

	var cmd Command
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		return err
	}

	c.log.Info("Recieved raft command", zap.Any("cmd", cmd))

	switch cmd.Operation {
	case SlotVsNodeChange:
		// Decode the data change for slot vs node change
		var cs ConfigSnapshot
		err = json.Unmarshal(cmd.Data, &cs)
		if err != nil {
			return err
		}

		c.handleSlotNodeChange(&cs)

	case AddRoute:
		var route rm.Route
		err := json.Unmarshal(cmd.Data, &route)
		if err != nil {
			return err
		}

		c.rStore.AddRoute(route.ID, &route)

	case RemoveRoute:
		var route rm.Route
		err := json.Unmarshal(cmd.Data, &route)
		if err != nil {
			return err
		}

		c.rStore.RemoveRoute(route.ID)
	}

	return nil
}

func (c *ConfigFSM) GetLastUpdatedTime() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastUpdateTime
}

func (c *ConfigFSM) SetChangeHandler(fn func() error) {
	c.onChangeHandler = fn
}

// Called when there is a change in node vs slot change.
// Assume that the state of node has changed and re-init everything
func (c *ConfigFSM) handleSlotNodeChange(cs *ConfigSnapshot) {

	// Re-initialise the DHT
	err := c.dht.Load(cs.Slots)
	if err != nil {
		c.log.Error("Unable to load dht in FSM", zap.Any("cs", cs), zap.Error(err))
		return
	}

	// Update the connections
	c.onChangeHandler()
}
