package server

import (
	"context"
	"fmt"
	"net"

	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/network"
	jobmodels "github.com/aarthikrao/timeMachine/models/jobmodels"
	"github.com/aarthikrao/timeMachine/process/cordinator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	network.JobStoreServer
	cp         jobstore.JobStoreWithReplicator
	grpcServer *grpc.Server
	log        *zap.Logger
}

func InitServer(
	cp *cordinator.CordinatorProcess,
	port int,
	log *zap.Logger,
) *server {

	jobStoreServer := &server{
		log: log,
		cp:  cp,
	}
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("failed to listen: %v", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	network.RegisterJobStoreServer(grpcServer, jobStoreServer)
	jobStoreServer.grpcServer = grpcServer

	go func() {
		log.Info("GRPC server started", zap.String("addr", addr))
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatal("Failed to listen", zap.Error(err))
		}
	}()

	return jobStoreServer
}

func (s *server) Close() {
	s.grpcServer.GracefulStop()
}

// GetJob fetches the job from a time machine instance
func (s *server) GetJob(ctx context.Context, jd *jobmodels.JobFetchDetails) (*jobmodels.JobCreationDetails, error) {
	job, err := s.cp.GetJob(jd.Collection, jd.ID)
	if err != nil {
		return nil, err
	}

	return &jobmodels.JobCreationDetails{
		ID:          job.ID,
		TriggerTime: int64(job.TriggerMS),
		Meta:        job.Meta,
		Route:       job.Route,
		Collection:  jd.Collection,
	}, err

}

// SetJob adds the job to a time machine instance
func (s *server) SetJob(ctx context.Context, jd *jobmodels.JobCreationDetails) (*jobmodels.JobCreationDetails, error) {
	_, err := s.cp.SetJob(jd.Collection, &jobmodels.Job{
		ID:        jd.ID,
		TriggerMS: int(jd.TriggerTime),
		Meta:      jd.Meta,
		Route:     jd.Route,
	})

	return jd, err
}

// DeleteJob will remove the job from time machine instance
func (s *server) DeleteJob(ctx context.Context, jd *jobmodels.JobFetchDetails) (*jobmodels.Empty, error) {
	_, err := s.cp.DeleteJob(jd.Collection, jd.ID)
	return &jobmodels.Empty{}, err
}

// ReplicateSetJob is the same as SetJob. It is called only by the leader to replicate the job on the follower
func (s *server) ReplicateSetJob(ctx context.Context, jd *jobmodels.JobCreationDetails) (*jobmodels.JobCreationDetails, error) {
	_, err := s.cp.ReplicateSetJob(jd.Collection, &jobmodels.Job{
		ID:        jd.ID,
		TriggerMS: int(jd.TriggerTime),
		Meta:      jd.Meta,
		Route:     jd.Route,
	})

	return jd, err
}

// ReplicateDeleteJob is the same as DeleteJobJob. It is called only by the leader to replicate the job on the follower
func (s *server) ReplicateDeleteJob(ctx context.Context, jd *jobmodels.JobFetchDetails) (*jobmodels.Empty, error) {
	_, err := s.cp.ReplicateDeleteJob(jd.Collection, jd.ID)
	return &jobmodels.Empty{}, err
}

// Health check
func (s *server) HealthCheck(context.Context, *jobmodels.HealthRequest) (*jobmodels.HealthResponse, error) {
	healthy, err := s.cp.HealthCheck()

	return &jobmodels.HealthResponse{
		Healthy: healthy,
	}, err
}
