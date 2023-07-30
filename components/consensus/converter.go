package consensus

import (
	"encoding/json"

	"github.com/aarthikrao/timeMachine/components/consensus/fsm"
	"github.com/aarthikrao/timeMachine/components/dht"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
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

func ConvertAddRoute(route *rm.Route) ([]byte, error) {
	by, err := json.Marshal(&route)
	if err != nil {
		return nil, err
	}

	cmd := fsm.Command{
		Operation: fsm.AddRoute,
		Data:      by,
	}

	return json.Marshal(&cmd)
}

func ConvertRemoveRoute(routeName string) ([]byte, error) {
	route := &rm.Route{
		ID: routeName,
		// Here, we are reusing the same route struct
		// cuz we lazy
	}

	by, err := json.Marshal(&route)
	if err != nil {
		return nil, err
	}

	cmd := fsm.Command{
		Operation: fsm.AddRoute,
		Data:      by,
	}

	return json.Marshal(&cmd)
}
