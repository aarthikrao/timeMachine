// Package jobstore contains the methods necessary to perform all operations on the DB.
// You can implement `JobStore` for data storage and retrieval over network or disk.
package jobstore

import (
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

// JobStoreType defines the type of underlying JobStore
type JobStoreType string

const (
	Database   JobStoreType = "db"
	WAL        JobStoreType = "wal"
	Network    JobStoreType = "network"
	Cordinator JobStoreType = "client"
)

// JobStore methods that are used to store and retrieve data across disk and network
type JobStore interface {
	GetJob(collection, jobID string) (*jm.Job, error)
	SetJob(collection string, job *jm.Job) (offset int64, err error)
	DeleteJob(collection, jobID string) (offset int64, err error)
}

// JobFetcher is used to fetch the jobs for executing them
type JobFetcher interface {
	JobStore

	// FetchJobForBucket is used to fetch all the jobs in the datastore in the given minute
	FetchJobForBucket(minute int) ([]*jm.Job, error)
	Close() error
}

// JobStoreWithReplicator adds replicate methods on top of JobStore interface
// This will be used by remote connections with need to replicate the data, check health etc.
type JobStoreWithReplicator interface {
	JobStore

	ReplicateSetJob(collection string, job *jm.Job) (offset int64, err error)
	ReplicateDeleteJob(collection, jobID string) (offset int64, err error)
	HealthCheck() (bool, error)
}
