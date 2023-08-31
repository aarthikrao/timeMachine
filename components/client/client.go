// This package is used to access the data layer from APIs like REST, GRPC and CLI
package client

import (
	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/routestore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"go.uber.org/zap"
)

type ClientProcess struct {
	nodeMgr *nodemanager.NodeManager

	rStore *routestore.RouteStore

	cp consensus.Consensus

	log *zap.Logger
}

// compile time validation
var _ jobstore.JobStore = &ClientProcess{}
var _ jobstore.JobFetcher = &ClientProcess{}

func CreateClientProcess(
	nodeMgr *nodemanager.NodeManager,
	rStore *routestore.RouteStore,
	cp consensus.Consensus,
	log *zap.Logger,
) *ClientProcess {
	return &ClientProcess{
		nodeMgr: nodeMgr,
		rStore:  rStore,
		cp:      cp,
		log:     log,
	}
}

// TODO: Add is leader check for route store.

func (cp *ClientProcess) GetJob(collection, jobID string) (*jm.Job, error) {
	slot, err := cp.nodeMgr.GetJobStoreInterface(jobID, true)
	if err != nil {
		return nil, err
	}
	cp.ownerCheck(slot)
	return slot.GetJob(collection, jobID)
}

func (cp *ClientProcess) SetJob(collection string, job *jm.Job) error {
	if collection == "" {
		return ErrInvalidDetails
	}

	if err := job.Valid(); err != nil {
		return err
	}

	slot, err := cp.nodeMgr.GetJobStoreInterface(job.ID, true)
	if err != nil {
		return err
	}
	cp.ownerCheck(slot)
	return slot.SetJob(collection, job)
}

func (cp *ClientProcess) DeleteJob(collection, jobID string) error {
	if collection == "" || jobID == "" {
		return ErrInvalidDetails
	}

	slot, err := cp.nodeMgr.GetJobStoreInterface(jobID, true)
	if err != nil {
		return err
	}
	cp.ownerCheck(slot)
	return slot.DeleteJob(collection, jobID)
}

func (cp *ClientProcess) Type() jobstore.JobStoreType {
	return jobstore.Client
}

// This should be used only for developement purpose
func (cp *ClientProcess) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	if minute == 0 {
		return nil, ErrInvalidDetails
	}

	return nil, nil // TODO: Yet to implement
}

func (cp *ClientProcess) GetRoute(routeID string) (*rm.Route, error) {
	if routeID == "" {
		return nil, ErrInvalidDetails
	}

	route := cp.rStore.GetRoute(routeID)
	if route == nil {
		return nil, ErrRouteNotFound
	}

	return route, nil
}

func (cp *ClientProcess) SetRoute(route *rm.Route) error {
	if err := route.Valid(); err != nil {
		return err
	}

	by, err := consensus.ConvertAddRoute(route)
	if err != nil {
		return err
	}

	// Update the consensus about the route
	return cp.cp.Apply(by)
}

func (cp *ClientProcess) DeleteRoute(routeID string) error {
	if routeID == "" {
		return ErrInvalidDetails
	}

	by, err := consensus.ConvertRemoveRoute(routeID)
	if err != nil {
		return err
	}

	// Delete the route from consensus
	return cp.cp.Apply(by)
}

// ownerCheck checks if the returned jobstore object is local, or on another time machine node
func (cp *ClientProcess) ownerCheck(slot jobstore.JobStore) {
	switch slot.Type() {
	case jobstore.Database, jobstore.WAL:
		break
	default:
		cp.log.Debug("Not the owner", zap.String("type", string(slot.Type())))
	}
}
