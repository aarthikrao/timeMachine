// This package is used to access the data layer from APIs like REST, GRPC and CLI
package client

import (
	ds "github.com/aarthikrao/timeMachine/components/datastore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
)

type ClientProcess struct {
	dataStore ds.DataStore
}

func CreateClientProcess(dataStore ds.DataStore) *ClientProcess {
	return &ClientProcess{
		dataStore: dataStore,
	}
}

func (cp *ClientProcess) GetJob(collection, jobID string) (*jm.Job, error) {
	return cp.dataStore.GetJob(collection, jobID)
}

func (cp *ClientProcess) SetJob(collection string, job *jm.Job) error {
	if collection == "" {
		return ErrInvalidDetails
	}

	if err := job.Valid(); err != nil {
		return err
	}

	return cp.dataStore.SetJob(collection, job)
}

func (cp *ClientProcess) DeleteJob(collection, jobID string) error {
	if collection == "" || jobID == "" {
		return ErrInvalidDetails
	}

	return cp.dataStore.DeleteJob(collection, jobID)
}

// This should be used only for developement purpose
func (cp *ClientProcess) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	if minute == 0 {
		return nil, ErrInvalidDetails
	}

	return cp.dataStore.FetchJobForBucket(minute)
}

func (cp *ClientProcess) GetRoute(routeID string) (*rm.Route, error) {
	if routeID == "" {
		return nil, ErrInvalidDetails
	}
	return cp.dataStore.GetRoute(routeID)
}

func (cp *ClientProcess) SetRoute(route *rm.Route) error {
	if err := route.Valid(); err != nil {
		return err
	}
	return cp.dataStore.SetRoute(route)

}

func (cp *ClientProcess) DeleteRoute(routeID string) error {
	if routeID == "" {
		return ErrInvalidDetails
	}
	return cp.dataStore.DeleteRoute(routeID)
}
