package network

import (
	"context"

	"github.com/aarthikrao/timeMachine/components/jobstore"
	"google.golang.org/grpc"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
)

type networkHandler struct {
	client JobStoreClient
	conn   *grpc.ClientConn
}

// Compile time interface validation
var _ jobstore.JobStore = &networkHandler{}

func CreateConnection(addr string) (*networkHandler, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock())

	if err != nil {
		return nil, err
	}

	return &networkHandler{
		client: NewJobStoreClient(conn),
		conn:   conn,
	}, nil
}

func (nh *networkHandler) GetJob(collection, jobID string) (*jm.Job, error) {
	resp, err := nh.client.GetJob(context.Background(), &jm.JobFetchDetails{
		ID:         jobID,
		Collection: collection,
	})

	if err != nil {
		return nil, err
	}

	return &jm.Job{
		TriggerTime: int(resp.TriggerTime),
		ID:          resp.ID,
		Meta:        resp.Meta,
		Route:       resp.Route,
	}, nil

}

func (nh *networkHandler) SetJob(collection string, job *jm.Job) error {
	_, err := nh.client.SetJob(context.Background(), &jm.JobCreationDetails{
		TriggerTime: int64(job.TriggerTime),
		ID:          job.ID,
		Meta:        job.Meta,
		Route:       job.Route,
		Collection:  collection,
	})

	return err
}

func (nh *networkHandler) DeleteJob(collection, jobID string) error {
	_, err := nh.client.DeleteJob(context.Background(), &jm.JobFetchDetails{
		Collection: collection,
		ID:         jobID,
	})

	return err
}

func (nh *networkHandler) Close() error {
	return nh.conn.Close() // TODO: Move all the connections to conn manager
}
