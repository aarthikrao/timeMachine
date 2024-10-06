// This package is used to access the data layer from APIs like REST, GRPC and CLI
package cordinator

import (
	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/executor"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/routestore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type CordinatorProcess struct {
	nodeMgr     *nodemanager.NodeManager
	rStore      *routestore.RouteStore
	dhtMgr      dht.DHT
	cp          consensus.Consensus
	selfNodeID  dht.NodeID
	jobExecutor executor.Executor
	log         *zap.Logger
}

// compile time validation
var _ jobstore.JobFetcher = (*CordinatorProcess)(nil)
var _ jobstore.JobStoreWithReplicator = (*CordinatorProcess)(nil)

func CreateCordinatorProcess(
	selfNodeID string,
	nodeMgr *nodemanager.NodeManager,
	rStore *routestore.RouteStore,
	cp consensus.Consensus,
	dhtMgr dht.DHT,
	jobExecutor executor.Executor,
	log *zap.Logger,
) *CordinatorProcess {
	return &CordinatorProcess{
		nodeMgr:     nodeMgr,
		rStore:      rStore,
		cp:          cp,
		dhtMgr:      dhtMgr,
		selfNodeID:  dht.NodeID(selfNodeID),
		jobExecutor: jobExecutor,
		log:         log,
	}
}

func (cp *CordinatorProcess) GetJob(collection, jobID string) (*jm.Job, error) {
	shardLoc, err := cp.dhtMgr.GetShard(jobID)
	if err != nil {
		return nil, err
	}

	shard, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return nil, err
	}

	if shard != nil {
		return shard.GetJob(collection, jobID)
	}

	// Local shard doesnt exist, fetch remote shard
	// TODO: Add read preferences from API
	conn, err := cp.nodeMgr.GetRemoteConnection(shardLoc.Leader)
	if err != nil {
		return nil, err
	}

	return conn.GetJob(collection, jobID)
}

func (cp *CordinatorProcess) SetJob(collection string, job *jm.Job) (offset int64, err error) {
	if collection == "" {
		return 0, ErrInvalidDetails
	}

	if err := job.Valid(); err != nil {
		return 0, err
	}

	shardLoc, err := cp.dhtMgr.GetShard(job.ID)
	if err != nil {
		return 0, err
	}

	if shardLoc.Leader != cp.selfNodeID {
		// This node is not the leader for this shard, hence we cannot serve write requests
		// Forward this request to the right owner
		remoteLeader, err := cp.nodeMgr.GetRemoteConnection(shardLoc.Leader)
		if err != nil {
			return 0, err
		}

		return remoteLeader.SetJob(collection, job)
	}

	// This means this node is the leader for this shard, we need to process the write request
	shard, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return 0, err
	}
	offset, err = shard.SetJob(collection, job)
	if err != nil {
		return 0, err
	}

	// Add the job to the executor queue
	if err = cp.jobExecutor.Queue(*job); err != nil {
		return 0, err
	}

	// Now we set the job in all the follower shards
	for _, follower := range shardLoc.Followers {
		remoteFollower, err := cp.nodeMgr.GetRemoteConnection(follower)
		if err != nil {
			return 0, err
		}

		_, err = remoteFollower.ReplicateSetJob(collection, job)
		if err != nil {
			return 0, err
		}
	}

	return offset, nil
}

func (cp *CordinatorProcess) DeleteJob(collection, jobID string) (offset int64, err error) {
	if collection == "" || jobID == "" {
		return 0, ErrInvalidDetails
	}

	shardLoc, err := cp.dhtMgr.GetShard(jobID)
	if err != nil {
		return 0, err
	}

	if shardLoc.Leader != cp.selfNodeID {
		// This node is not the leader for this shard, hence we cannot serve write requests
		// Forward this request to the right owner
		remoteLeader, err := cp.nodeMgr.GetRemoteConnection(shardLoc.Leader)
		if err != nil {
			return 0, err
		}

		offset, err = remoteLeader.DeleteJob(collection, jobID)
		if err != nil {
			return 0, err
		}

		// No more processing in this node
		return offset, nil
	}

	// This means this node is the leader for this shard, we need to process the write request
	shard, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return 0, err
	}
	offset, err = shard.DeleteJob(collection, jobID)
	if err != nil {
		return 0, err
	}

	// Now we set the job in all the follower shards
	for _, follower := range shardLoc.Followers {
		remoteFollower, err := cp.nodeMgr.GetRemoteConnection(follower)
		if err != nil {
			return 0, err
		}

		_, err = remoteFollower.ReplicateDeleteJob(collection, jobID)
		if err != nil {
			return 0, err
		}
	}

	return offset, nil
}

// ReplicateSetJob can be only called from the master
func (cp *CordinatorProcess) ReplicateSetJob(collection string, job *jm.Job) (offset int64, err error) {
	shardLoc, err := cp.dhtMgr.GetShard(job.ID)
	if err != nil {
		return 0, err
	}

	localFollowerSlot, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return 0, errors.Wrap(err, "follower slot: ")
	}

	if offset, err := localFollowerSlot.SetJob(collection, job); err != nil {
		return offset, errors.Wrap(err, "follower slot: ")
	}

	return offset, nil
}

func (cp *CordinatorProcess) ReplicateDeleteJob(collection, jobID string) (offset int64, err error) {
	shardLoc, err := cp.dhtMgr.GetShard(jobID)
	if err != nil {
		return 0, err
	}

	localFollowerSlot, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return 0, errors.Wrap(err, "follower slot: ")
	}

	if offset, err := localFollowerSlot.DeleteJob(collection, jobID); err != nil {
		return offset, errors.Wrap(err, "follower slot: ")
	}

	return offset, nil
}

func (cp *CordinatorProcess) Type() jobstore.JobStoreType {
	return jobstore.Cordinator
}

// This should be used only for developement purpose
func (cp *CordinatorProcess) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	if minute == 0 {
		return nil, ErrInvalidDetails
	}

	return nil, nil // TODO: Yet to implement
}

func (cp *CordinatorProcess) GetRoute(routeID string) (*rm.Route, error) {
	if routeID == "" {
		return nil, ErrInvalidDetails
	}

	route := cp.rStore.GetRoute(routeID)
	if route == nil {
		return nil, ErrRouteNotFound
	}

	return route, nil
}

func (cp *CordinatorProcess) SetRoute(route *rm.Route) error {
	if err := route.Valid(); err != nil {
		return err
	}

	by, err := consensus.ConvertAddRoute(route)
	if err != nil {
		return err
	}

	// Update the consensus about the route
	return cp.cp.Apply(by)
}

func (cp *CordinatorProcess) DeleteRoute(routeID string) error {
	if routeID == "" {
		return ErrInvalidDetails
	}

	by, err := consensus.ConvertRemoveRoute(routeID)
	if err != nil {
		return err
	}

	// Delete the route from consensus
	return cp.cp.Apply(by)
}

func (cp *CordinatorProcess) HealthCheck() (bool, error) {
	return true, nil // We are ready to accept new requests. So we always return true
}

// Dummy method to satisfy the JobFetcher interface. Client will not usually call this method.
func (cp *CordinatorProcess) Close() error {
	return nil

}
