# Time Machine DB 🐓
[![Discord](https://img.shields.io/badge/Discord-%235865F2.svg?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/pDGNPj3dTM) 
![Status](https://img.shields.io/badge/Status-Ideation-ffb3ff?style=for-the-badge)

A distributed, fault tolerant scheduler database that can potentially scale to millions of jobs. 

The idea is to build it with a storage layer based on B+tree or LSM-tree implementation, consistent hashing for load balancing, and raft for consensus.

## 🧬 Documentation
- [Purpose](./docs/Purpose.md)
- [Architecture](./docs/Architecture.md) • [Components of a node](/components/Components.md) • [Also read](./docs/Refer.md)
- [Developer APIs](./docs/DevAPI.md) • [Job APIs](./docs/DevAPI.md#-job-apis) • [Route APIs](./docs/DevAPI.md#-route-apis)
- [TODO](./docs/TODO.md)

![Cluster animation](/docs/images/cluster_animation.gif)

## 🎯 Quick start

```bash
# Start 3 nodes. Create respective data folders 
# as data/node1/data and data/node1/raft
./create-cluster.sh 3

# To add node2 to node1 as to form cluster
❯ curl -X POST 'http://localhost:8001/cluster/join' \
-H 'Content-Type: application/json' \
--data-raw '{
 "node_id":"node2",
 "raft_address":"localhost:8102"
}'
# Do the same thing to node3.

# To check cluster health of 3 nodes
❯ ./check-health.sh 3

# More scripts coming soon
```

## 🎬 Roadmap
- [x] Core project structure
- [x] Data storage layer
    - [x] Implement BoltDB
    - [ ] Optimise to Messagepack, proto or avro
- [x] Bash/Make script
    - [x] Cluster deployment
    - [x] Build and run tests
    - [ ] Add and remove nodes
- [x] Client CRUD
    - [x] Rest interface
- [x] Node leader election
    - [x] Implement Raft
    - [x] Implement FSM
    - [x] Add/Remove nodes
- [ ] `vnode` leader election
    - [ ] Failure and restart
    - [ ] `vnode` leader and follower health check
- [ ] Node connection manager
    - [ ] GRPC contracts and message passing
    - [ ] Data replication
- [ ] Properties file
    - [ ] Validation
    - [ ] Using master properties file
- [x] Partioner Hash function
    - [x] Hashring algorithm
    - [x] Adding and removing nodes
    - [ ] Provision for clustering key
    - [ ] Re-routing via connection manager
- [ ] Restart, scale up and scale down handling
    - [ ] Invoking node and `vnode` leader election
- [ ] Job executor
    - [ ] Hybrid logical clock
    - [ ] Rest caller
    - [ ] Concensus during publish

## 🛺 Tech Stack
Refer [Tech stack](/docs/Refer.md#🛺-tech-stack)

## ⚽ Contribute
* Choose a component to work on.
* Research the component thoroughly.
* Reach out to me, so that I can mark it as "Work in Progress" to prevent duplication of efforts.
* Build, code, and test the component.
* Submit a pull request (PR) when you are ready to have your changes reviewed.


Refer [Contributing](./CONTRIBUTING.md) for more