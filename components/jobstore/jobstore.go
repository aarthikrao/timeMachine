// jobstore contains the methods necessary to perform all operations on the DB
// You can implement `JobStore` for data storage and retrieval over network or disk
package jobstore

import (
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

// JobStore methods that are used to store and retrieve data across disk and network
type JobStore interface {
	GetJob(collection, jobID string) (*jm.Job, error)
	SetJob(collection string, job *jm.Job) error
	DeleteJob(collection, jobID string) error
}

// JobFetcher is used to fetch the jobs for executing them
type JobFetcher interface {
	JobStore

	// FetchJobTill is used to fetch all the jobs in the datastore till the provided time
	FetchJobForBucket(minute int) ([]*jm.Job, error)
}

type JobStoreConn interface {
	JobStore
	Close() error
}
