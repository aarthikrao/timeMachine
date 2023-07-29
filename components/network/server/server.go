package server

import (
	"context"
	"fmt"
	"net"

	"github.com/aarthikrao/timeMachine/components/client"
	"github.com/aarthikrao/timeMachine/components/network"
	jobmodels "github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	network.JobStoreServer
	cp         *client.ClientProcess
	grpcServer *grpc.Server
	log        *zap.Logger
}

func InitServer(
	cp *client.ClientProcess,
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
	err := s.cp.SetJob(jd.Collection, &jobmodels.Job{
		ID:        jd.ID,
		TriggerMS: int(jd.TriggerTime),
		Meta:      jd.Meta,
		Route:     jd.Route,
	})

	return jd, err
}

// DeleteJob will remove the job from time machine instance
func (s *server) DeleteJob(ctx context.Context, jd *jobmodels.JobFetchDetails) (*jobmodels.Empty, error) {
	return &jobmodels.Empty{}, s.cp.DeleteJob(jd.Collection, jd.ID)
}
