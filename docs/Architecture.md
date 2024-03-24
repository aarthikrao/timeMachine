# 🔮 Architecture
We want to build a scalable, consistent, fault tolerant scheduler that is simple to use. Time machine is inspired by many projects
* [ScyllaDB](https://github.com/scylladb/scylladb) for ring architecture
* [CockroachDB](https://github.com/cockroachdb/cockroach) for consistency and scalability. It is written in Go and its open source
* [Redis](redis.io) for its simplicity. We want to support REdis Serialization Protocol (RESP) in the future.

## ⚙️ Design choices
- **Node Discovery, Failure Detection, Membership Management**: We use the Raft consensus algorithm. Raft helps us manage cluster membership and detect failures efficiently, ensuring high availability.
- **Data Replication, Consistency**: Our system is based on a partitioned master-slave architecture. It defaults to strong consistency, but allows for tunable consistency levels according to application needs.
- **Load Balancing, Data Partitioning**: We implement the hash shard algorithm for effective data partitioning and load balancing, optimizing resource utilization and response times.
- **Rebalancing**: Due to its critical nature, rebalancing is performed manually.  These operations are supposed to be executed with care under low-traffic conditions
- **Query Interface**: Currently, we offer a REST API for queries. Plans are in place to support the Redis Serialization Protocol (RESP) in the future, catering to more use cases and improving efficiency.
- **Storage**: We chose BBoltDB for its B-tree based implementation. This choice suits our need for efficient range scans. We're open to incorporating LSM based storage solutions in the future to further enhance our system's capabilities.
- **Encoding**: Data is stored in binary format, preferably using protobuf. This method offers compact storage and fast serialization/deserialization, contributing to overall system performance.


## 🦋 Data distribution
Refer [DHT component](./../components/dht/dht.md)

## ✏️ Definitions
| Term                        | Description                                                                                                                                                                         |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Admin                       | A human with special privileges to manually trigger operations like compaction, rebalance, etc., under low-traffic conditions. Robots are not permitted.                            |
| Cluster                     | A set of timeMachine nodes that work together.                                                                                                                                      |
| Collection                  | Logical grouping of jobs.                                                                                                                                                           |
| Creator                     | The service which created the job.                                                                                                                                                  |
| DHT                         | Distributed Hash Table. A decentralized distributed system that provides a lookup service similar to a hash table; keys are assigned to nodes making data location efficient.       |
| Failure Detection           | The mechanism by which the system identifies and handles node failures.                                                                                                             |
| Hash function               | A function that can be used to map data of arbitrary size to fixed-size values. Used to determine the placement of jobs in shards within the system.                                |
| Job                         | A job holds information related to a callback, triggered at `trigger_time` and contains `meta` information passed on to the `recipient`.                                            |
| Load Balancing              | Distributing workloads across multiple nodes to ensure efficient processing and to avoid overloading any single node.                                                               |
| Membership Management       | Managing the list of active nodes and their roles within the cluster.                                                                                                               |
| Node                        | A physical server instance of timeMachine Denoted by `NodeID`. |
| Node-leader                 | The leader node, selected based on the Raft consensus algorithm.                                                                                                                    |
| Raft Consensus              | A consensus algorithm used for managing a replicated log and ensuring data consistency across cluster nodes.                                                                        |
| Recipient                   | The service that receives the trigger. Currently, only REST Webhooks are supported.                                                                                                 |
| Replica                     | A copy of the shard. The number of replicas is specified when initializing the time machine cluster, distributed to ensure a single copy of a job per node.                         |
| Replicated Sharded Architecture | A system topology where data is partitioned into shards for scalability and each shard is replicated across multiple nodes for fault tolerance and high availability. This ensures that even if a node fails, the system can continue to operate seamlessly by serving data from replica shards. |
| RPC                         | Remote Procedure Call. A protocol that one program can use to request a service from a program located on another computer on a network without having to understand network details. |
| Route                       | The routing information used to publish a job. Multiple routes can be created for callbacks.                                                                                        |
| Shard                       | A subset of the collection stored in sorted order. A job's resident shard is determined by hashing the `job_id`. Identified by `ShardID`.A physical node will be assigned multiple non-contiguous slots. In case of scale-up or down, slots as a whole will be transferred over to the new nodes. This will make it easier to rebalance a timeMachine cluster. Refer [DHT](../components/dht/dht.md)                                                                  |
| Shard-leader                | The leader shard that handles creation, deletion, and triggering for all jobs in that shard.                                                                                        |
| Trigger exactly once semantic | To make sure that a job is triggered exactly once, timeMachine elects a leader for every shard. This shard-leader is responsible to trigger the job and publish the offset location for this particular shard. One physical node may contain multiple shard leaders. The shard leaders are distributed such that all the physical nodes get an equal number of shard-leaders. This maked sure the trigger workload is spread throughout the entire cluster. In case a shard leader fails, a follower shard becomes the leader and takes the responsibility of publishing the job from the latest committed offset.                                                 |
| Tunable Consistency         | The ability to adjust the level of consistency (e.g., strong, eventual) based on specific requirements.                                                                             |
| Webhook                     | A method of augmenting or altering the behavior of a web page or web application with custom callbacks. Used to receive triggers in a RESTful way.                                  |


## 🎰 Behavior

### Creating and Triggering Jobs

Jobs can be created through our [Developer APIs](./DevAPI.md#create-a-job) by specifying a `trigger_time`, `route`, and optionally, `meta` information. Once created, the job is replicated across all shard replicas, including the leader and its followers, to ensure reliability and fault tolerance.

When the specified `trigger_time` is reached, the shard leader retrieves the job and triggers it by sending a POST request to the webhook URL specified in the `route`. This ensures that the job is executed exactly at its scheduled time, maintaining the trigger's precision and reliability.

### Routing Webhooks

Routes define how and where a job's callback is delivered. To establish a route, use the [Developer APIs](./DevAPI.md#create-a-route). At the job's `trigger_time`, timeMachine issues a POST request to the configured webhook URL in the route. This mechanism allows for flexible and dynamic job routing, enabling targeted job execution across diverse endpoints.

### Hash Partitioning Algorithm

To efficiently distribute and locate jobs across the cluster, timeMachine DB employs the [xxHash](https://cyan4973.github.io/xxHash/) algorithm. Jobs are assigned to shards based on the hash of their `job_id` using the following formula:

```plaintext
Shard Number = xxHash(job_id) % number_of_shards
```
The location of the node for a key is derived from the dht. You can read more about this in the [DHT component](/components/dht/dht.md)

### RPCs and message passing
refer [MessagePassing](./MessagePassing.md)

## 💡 Inspirations
* [Designing Data-Intensive Applications](https://www.oreilly.com/library/view/designing-data-intensive-applications/9781491903063/)
* [ScyllaDB](https://github.com/scylladb/scylladb)
* [Raft consensus algorithm](https://raft.github.io/)