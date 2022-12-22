// datastore contains the methods necessary to perform all operations on the DB
// You can implement `JobStore` and `RouteStore` to use any type of embeded storage
package datastore

import (
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"

	rm "github.com/aarthikrao/timeMachine/models/routemodels"
)

type DataStore interface {
	GetJob(collection, jobID string) (*jm.Job, error)
	SetJob(collection string, job *jm.Job) error
	DeleteJob(collection, jobID string) error

	// FetchJobTill is used to fetch all the jobs in the datastore till the provided time
	FetchJobForBucket(minute int) ([]*jm.Job, error)

	GetRoute(routeID string) (*rm.Route, error)
	SetRoute(route *rm.Route) error
	DeleteRoute(route string) error
	Close() error
}
