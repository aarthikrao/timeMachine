package concensus

import (
	"encoding/json"

	"github.com/aarthikrao/timeMachine/components/concensus/fsm"
	"github.com/aarthikrao/timeMachine/components/dht"
)

func ConvertConfigSnapshot(Slots map[dht.SlotID]*dht.SlotInfo) ([]byte, error) {
	cs := fsm.ConfigSnapshot{
		Slots: Slots,
	}
	by, err := json.Marshal(&cs)
	if err != nil {
		return nil, err
	}

	cmd := fsm.Command{
		Operation: fsm.SlotVsNodeChange,
		Data:      by,
	}

	return json.Marshal(&cmd)
}
