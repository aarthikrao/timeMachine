// This package is used to access the data layer from APIs like REST, GRPC and CLI
package cordinator

import (
	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/components/routestore"
	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type CordinatorProcess struct {
	nodeMgr *nodemanager.NodeManager

	rStore *routestore.RouteStore

	dhtMgr dht.DHT

	cp consensus.Consensus

	selfNodeID dht.NodeID

	log *zap.Logger
}

// compile time validation
var _ jobstore.JobStore = (*CordinatorProcess)(nil)
var _ jobstore.JobFetcher = (*CordinatorProcess)(nil)

func CreateCordinatorProcess(
	selfNodeID string,
	nodeMgr *nodemanager.NodeManager,
	rStore *routestore.RouteStore,
	cp consensus.Consensus,
	dhtMgr dht.DHT,
	log *zap.Logger,
) *CordinatorProcess {
	return &CordinatorProcess{
		nodeMgr:    nodeMgr,
		rStore:     rStore,
		cp:         cp,
		dhtMgr:     dhtMgr,
		selfNodeID: dht.NodeID(selfNodeID),
		log:        log,
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

func (cp *CordinatorProcess) SetJob(collection string, job *jm.Job) error {
	if collection == "" {
		return ErrInvalidDetails
	}

	if err := job.Valid(); err != nil {
		return err
	}

	shardLoc, err := cp.dhtMgr.GetShard(job.ID)
	if err != nil {
		return err
	}

	if shardLoc.Leader != cp.selfNodeID {
		// This node is not the leader for this shard, hence we cannot serve write requests
		// Forward this request to the right owner
		conn, err := cp.nodeMgr.GetRemoteConnection(shardLoc.Leader)
		if err != nil {
			return err
		}

		err = conn.SetJob(collection, job)
		if err != nil {
			return err
		}

		// No more processing in this node
		return nil
	}

	// This means this node is the leader for this shard, we need to process the write request
	shard, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return err
	}
	err = shard.SetJob(collection, job)
	if err != nil {
		return err
	}

	// Now we set the job in all the follower shards
	for _, follower := range shardLoc.Followers {
		conn, err := cp.nodeMgr.GetRemoteConnection(follower)
		if err != nil {
			return err
		}

		err = conn.ReplicateSetJob(collection, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cp *CordinatorProcess) DeleteJob(collection, jobID string) error {
	if collection == "" || jobID == "" {
		return ErrInvalidDetails
	}

	shardLoc, err := cp.dhtMgr.GetShard(jobID)
	if err != nil {
		return err
	}

	if shardLoc.Leader != cp.selfNodeID {
		// This node is not the leader for this shard, hence we cannot serve write requests
		// Forward this request to the right owner
		conn, err := cp.nodeMgr.GetRemoteConnection(shardLoc.Leader)
		if err != nil {
			return err
		}

		err = conn.DeleteJob(collection, jobID)
		if err != nil {
			return err
		}

		// No more processing in this node
		return nil
	}

	// This means this node is the leader for this shard, we need to process the write request
	shard, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return err
	}
	err = shard.DeleteJob(collection, jobID)
	if err != nil {
		return err
	}

	// Now we set the job in all the follower shards
	for _, follower := range shardLoc.Followers {
		conn, err := cp.nodeMgr.GetRemoteConnection(follower)
		if err != nil {
			return err
		}

		err = conn.ReplicateDeleteJob(collection, jobID)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReplicateSetJob can be only called from the master
func (cp *CordinatorProcess) ReplicateSetJob(collection string, job *jm.Job) error {
	shardLoc, err := cp.dhtMgr.GetShard(job.ID)
	if err != nil {
		return err
	}

	localFollowerSlot, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	if err := localFollowerSlot.SetJob(collection, job); err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	return nil
}

func (cp *CordinatorProcess) ReplicateDeleteJob(collection, jobID string) error {
	shardLoc, err := cp.dhtMgr.GetShard(jobID)
	if err != nil {
		return err
	}

	localFollowerSlot, err := cp.nodeMgr.GetLocalShard(shardLoc.ID)
	if err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	if err := localFollowerSlot.DeleteJob(collection, jobID); err != nil {
		return errors.Wrap(err, "follower slot: ")
	}

	return nil
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

// Dummy method to satisfy the JobFetcher interface. Client will not usually call this method.
func (cp *CordinatorProcess) Close() error {
	return nil

}
