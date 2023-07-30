# 🎬 Roadmap

## Phase 1
- [x] Core project structure
- [x] Data storage layer
    - [x] Implement BoltDB
    - [ ] Optimise to Messagepack, proto or avro
- [x] Bash/Make script
    - [x] Cluster deployment
    - [x] Build and run tests
    - [x] Add and remove nodes
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
    - [ ] Consensus during publish

## Phase 2
TBD