package network

import (
	"context"
	"time"

	"github.com/aarthikrao/timeMachine/components/jobstore"
	"google.golang.org/grpc"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

type networkHandler struct {
	client     JobStoreClient
	conn       *grpc.ClientConn
	rpcTimeout time.Duration
}

// Compile time interface validation
var _ jobstore.JobStoreWithReplicator = (*networkHandler)(nil)

func CreateJobStoreClient(conn *grpc.ClientConn, rpcTimeout time.Duration) *networkHandler {
	return &networkHandler{
		client:     NewJobStoreClient(conn),
		conn:       conn,
		rpcTimeout: rpcTimeout,
	}
}

func (nh *networkHandler) GetJob(collection, jobID string) (*jm.Job, error) {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(nh.rpcTimeout))
	defer cancelFunc()

	resp, err := nh.client.GetJob(ctx, &jm.JobFetchDetails{
		ID:         jobID,
		Collection: collection,
	})

	if err != nil {
		return nil, err
	}

	return &jm.Job{
		TriggerMS: int(resp.TriggerTime),
		ID:        resp.ID,
		Meta:      resp.Meta,
		Route:     resp.Route,
	}, nil

}

func (nh *networkHandler) SetJob(collection string, job *jm.Job) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(nh.rpcTimeout))
	defer cancelFunc()

	_, err := nh.client.SetJob(ctx, &jm.JobCreationDetails{
		TriggerTime: int64(job.TriggerMS),
		ID:          job.ID,
		Meta:        job.Meta,
		Route:       job.Route,
		Collection:  collection,
	})

	return err
}

func (nh *networkHandler) DeleteJob(collection, jobID string) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(nh.rpcTimeout))
	defer cancelFunc()

	_, err := nh.client.DeleteJob(ctx, &jm.JobFetchDetails{
		Collection: collection,
		ID:         jobID,
	})

	return err
}

func (nh *networkHandler) Type() jobstore.JobStoreType {
	return jobstore.Network
}

func (nh *networkHandler) ReplicateSetJob(collection string, job *jm.Job) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(nh.rpcTimeout))
	defer cancelFunc()

	_, err := nh.client.ReplicateSetJob(ctx, &jm.JobCreationDetails{
		TriggerTime: int64(job.TriggerMS),
		ID:          job.ID,
		Meta:        job.Meta,
		Route:       job.Route,
		Collection:  collection,
	})

	return err
}

func (nh *networkHandler) ReplicateDeleteJob(collection, jobID string) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(nh.rpcTimeout))
	defer cancelFunc()

	_, err := nh.client.DeleteJob(ctx, &jm.JobFetchDetails{
		Collection: collection,
		ID:         jobID,
	})

	return err
}
