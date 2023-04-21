// This package is used to access the data layer from APIs like REST, GRPC and CLI
package client

import (
	"github.com/aarthikrao/timeMachine/components/jobstore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
)

type ClientProcess struct {
	nodeMgr *nodemanager.NodeManager
}

// compile time validation
var _ jobstore.JobStore = &ClientProcess{}

func CreateClientProcess(nodeMgr *nodemanager.NodeManager) *ClientProcess {
	return &ClientProcess{
		nodeMgr: nodeMgr,
	}
}

func (cp *ClientProcess) GetJob(collection, jobID string) (*jm.Job, error) {
	slot, err := cp.nodeMgr.GetLocation(jobID)
	if err != nil {
		return nil, err
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

	slot, err := cp.nodeMgr.GetLocation(job.ID)
	if err != nil {
		return err
	}
	return slot.SetJob(collection, job)
}

func (cp *ClientProcess) DeleteJob(collection, jobID string) error {
	if collection == "" || jobID == "" {
		return ErrInvalidDetails
	}

	slot, err := cp.nodeMgr.GetLocation(jobID)
	if err != nil {
		return err
	}
	return slot.DeleteJob(collection, jobID)
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

	return nil, nil // TODO: Yet to implement
}

func (cp *ClientProcess) SetRoute(route *rm.Route) error {
	if err := route.Valid(); err != nil {
		return err
	}

	return nil // TODO: Yet to implement

}

func (cp *ClientProcess) DeleteRoute(routeID string) error {
	if routeID == "" {
		return ErrInvalidDetails
	}

	return nil // TODO: Yet to implement
}
