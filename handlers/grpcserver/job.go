package grpcserver

import (
	"context"

	"github.com/aarthikrao/timeMachine/components/client"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/network"
	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type jobServer struct {
	network.JobStoreServer

	// A job client that implements the JobStore interface
	jobClient jobstore.JobStore
	log       *zap.Logger
}

// CreateJobServer registers the job server to the provided grpc server.
// It can then be used to fetch the jobs and details from this server
func CreateJobServer(
	grpcServer *grpc.Server,
	jobClient *client.ClientProcess,
	log *zap.Logger,
) *jobServer {

	log.Info("starting grpc job server")
	js := &jobServer{
		log: log,
	}

	network.RegisterJobStoreServer(grpcServer, js)
	return js
}

// GetJob fetches the job from a time machine instance
func (js *jobServer) GetJob(ctx context.Context, jobDetails *jobmodels.JobFetchDetails) (*jobmodels.JobCreationDetails, error) {
	job, err := js.jobClient.GetJob(jobDetails.Collection, jobDetails.ID)
	if err != nil {
		return nil, err
	}

	return &jobmodels.JobCreationDetails{
		ID:          job.ID,
		Meta:        job.Meta,
		Route:       job.Route,
		Collection:  jobDetails.Collection,
		TriggerTime: int64(job.TriggerMS),
	}, nil
}

// SetJob adds the job to a time machine instance
func (js *jobServer) SetJob(ctx context.Context, jobDetails *jobmodels.JobCreationDetails) (*jobmodels.JobCreationDetails, error) {
	return jobDetails, js.jobClient.SetJob(jobDetails.Collection, &jobmodels.Job{ // TODO: Do we need to send the job creation details again
		ID:        jobDetails.ID,
		TriggerMS: int(jobDetails.TriggerTime),
		Meta:      jobDetails.Meta,
		Route:     jobDetails.Route,
	})

}

// DeleteJob will remove the job from time machine instance
func (js *jobServer) DeleteJob(ctx context.Context, jobDetails *jobmodels.JobFetchDetails) (*jobmodels.Empty, error) {
	return &jobmodels.Empty{},
		js.jobClient.DeleteJob(jobDetails.Collection, jobDetails.ID)
}
