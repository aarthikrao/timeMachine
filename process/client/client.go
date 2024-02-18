// This package is used to access the data layer from APIs like REST, GRPC and CLI
package client

import (
	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/routestore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ClientProcess struct {
	nodeMgr *nodemanager.NodeManager

	rStore *routestore.RouteStore

	cp consensus.Consensus

	log *zap.Logger
}

// compile time validation
var _ jobstore.JobFetcher = (*ClientProcess)(nil)
var _ jobstore.JobStoreWithReplicator = (*ClientProcess)(nil)

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
	var slot jobstore.JobStore

	owned, err := cp.nodeMgr.IsSlotOwner(jobID, true)
	if err != nil {
		return nil, err
	}

	if owned {
		slot, err = cp.nodeMgr.GetLocalSlot(jobID, true)
		if err != nil {
			return nil, err
		}

	} else {
		slot, err = cp.nodeMgr.GetRemoteSlot(jobID, true)
		if err != nil {
			return nil, err
		}

	}

	return slot.GetJob(collection, jobID)
}

func (cp *ClientProcess) SetJob(collection string, job *jm.Job) error {
	if collection == "" {
		return ErrInvalidDetails
	}

	if err := job.Valid(); err != nil {
		return err
	}

	owned, err := cp.nodeMgr.IsSlotOwner(job.ID, true)
	if err != nil {
		return err
	}

	if !owned {
		// Forward the request to owner node
		slot, err := cp.nodeMgr.GetRemoteSlot(job.ID, true)
		if err != nil {
			return err
		}

		return slot.SetJob(collection, job)
	}

	// add job to the local slot
	localSlot, err := cp.nodeMgr.GetLocalSlot(job.ID, true)
	if err != nil {
		return err
	}

	if err := localSlot.SetJob(collection, job); err != nil {
		return err
	}

	// add job to the follower
	remoteFollowerSlot, err := cp.nodeMgr.GetRemoteSlot(job.ID, false)
	if err != nil {
		return err
	}

	return remoteFollowerSlot.ReplicateSetJob(collection, job)
}

func (cp *ClientProcess) DeleteJob(collection, jobID string) error {
	if collection == "" || jobID == "" {
		return ErrInvalidDetails
	}

	owned, err := cp.nodeMgr.IsSlotOwner(jobID, true)
	if err != nil {
		return err
	}

	if !owned {
		// Forward the request to owner node
		slot, err := cp.nodeMgr.GetRemoteSlot(jobID, true)
		if err != nil {
			return err
		}

		return slot.DeleteJob(collection, jobID)
	}

	// Update the local slot
	localSlot, err := cp.nodeMgr.GetLocalSlot(jobID, true)
	if err != nil {
		return err
	}

	if err := localSlot.DeleteJob(collection, jobID); err != nil {
		return err
	}

	// Update the follower
	remoteFollowerSlot, err := cp.nodeMgr.GetRemoteSlot(jobID, false)
	if err != nil {
		return err
	}

	return remoteFollowerSlot.ReplicateDeleteJob(collection, jobID)
}

// ReplicateSetJob can be only called from the master
func (cp *ClientProcess) ReplicateSetJob(collection string, job *jm.Job) error {
	localFollowerSlot, err := cp.nodeMgr.GetLocalSlot(job.ID, false)
	if err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	if err := localFollowerSlot.SetJob(collection, job); err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	return nil
}

func (cp *ClientProcess) ReplicateDeleteJob(collection, jobID string) error {
	localFollowerSlot, err := cp.nodeMgr.GetLocalSlot(jobID, false)
	if err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	if err := localFollowerSlot.DeleteJob(collection, jobID); err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	return nil
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

func (cp *ClientProcess) HealthCheck() (bool, error) {
	return true, nil // We are ready to accept new requests. So we always return true
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

// Dummy method to satisfy the JobFetcher interface. Client will not usually call this method.
func (cp *ClientProcess) Close() error {
	return nil
}
