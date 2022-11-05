# Architecture
Time machine follows a Scylla-like ring architecture which is inspired by Apache Cassandra, Amazon Dynamo and Google BigTable.

## ✏️ Definitions
### Terms
* `database` - A set of collections
* `collection` - Logical grouping of jobs
* `job` - A job holds information related to a task. This job is triggered at the `trigger_time`. It also contains `meta` information that can be saved while creating the job. This is passed on to the `recipient` at the trigger time.

### Components
* `cluster` - A set of timeMachine nodes that work together
* `node` - A physical server instance of timeMachine
* `vnode` - A subset of the collection stored in sorted order. The resident vnode of a `job` can be found by using the hash function on the `job_id`. Refer [vnode](#vnode)

### Actors
* `creator` - the service which created this job
* `recipient` - the service which receives the trigger. Currently, only REST Webhook is supported
* `vnode-leader` - the leader elected to trigger jobs for a particular time slot.
* `admin` - a human who has special privileges to manually trigger compaction, rebalance etc. These operations are supposed to be executed with care under low-traffic conditions. I'm sorry, robots aren't allowed.

## 🎰 Behaviour
### vnode
A vnode is a subset of a collection of jobs that are stored in sorted order. Overall, a collection will have a `num_vnodes` number of vnodes evenly spread across the cluster. The default value is 12 `(LCM{3,4})`. A physical node will be assigned multiple non-contiguous Vnodes. In case of scale-up or down, vnodes as a whole will be transferred over to the new nodes. This will make it easier to rebalance a timeMachine cluster.

### Trigger exactly once semantic
To make sure that a job is triggered exactly once, timeMachine elects a leader for every vnode for a time slot. This vnode-leader is responsible to trigger the job and publish the offset location for this particular vnode. One physical node may contain multiple vnode leaders. In case a vnode leader fails, a follower vnode becomes the leader and takes the responsibility of publishing the job from the latest committed offset. For leader election, timeMachine will use [Raft consensus algorithm](https://raft.github.io/). #TODO: Revisit

### Creating a job
When a job is saved in timeMachine, a [hash partitioner algorithm](#hash-partitioner-algorithm)(with `job_id` as an input) determines the vnode number to which this job will be assigned. The job will be published to all the nodes that contain the vnode (both followers and leaders). In case of conflicts, we will follow [Conflict resolution](#conflict-resolution)

### Conflict resolution
To be done

### Hash partitioner algorithm
To be done

## 💡 Inspirations
* [Designing Data-Intensive Applications](https://www.oreilly.com/library/view/designing-data-intensive-applications/9781491903063/)
* [ScyllaDB](https://github.com/scylladb/scylladb)
* [Raft consensus algorithm](https://raft.github.io/)