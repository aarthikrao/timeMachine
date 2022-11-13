# Time Machine DB 🐓
[![Discord](https://img.shields.io/badge/Discord-%235865F2.svg?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/pDGNPj3dTM) 
![Status](https://img.shields.io/badge/Status-Ideation-ffb3ff?style=for-the-badge)

A distributed, fault tolerant scheduler database that can potentially scale to millions of jobs. 

The idea is to build it with a storage layer based on B+tree or LSM-tree implementation, consistent hashing for load balancing, and raft for consensus.

## 🧬 Documentation
- [Purpose](./docs/Purpose.md)
    - [Why are we building this ?](./docs/Purpose.md#-why-are-we-building-this)
    - [What does it take ?](./docs/Purpose.md#-what-does-it-take)
    - [What this is not ?](./docs/Purpose.md#-what-this-is-not)
- [Architecture](./docs/Architecture.md)
- [Developer APIs](./docs/DevAPI.md)
    - [Job APIs](./docs/DevAPI.md#-job-apis)
    - [Route APIs](./docs/DevAPI.md#-route-apis)
- [TODO](./docs/TODO.md)

## 🎬 Roadmap
- [x] *Under development*
- [ ] *Yet to be picked*


- [x] Core project structure
- [x] Data storage layer
    - [x] Implement BoltDB
    - [ ] Implement Badger
    - [ ] Optimise to Messagepack, proto or avro
- [x] Client CRUD
    - [x] Rest interface
    - [ ] GRPC Interface
- [x] Node leader election
    - [x] Implement Raft
    - [x] Implement FSM
    - [x] Add/Remove nodes
- [ ] `vnode` leader election
- [ ] Node connection manager
    - [ ] Message passing
    - [ ] Data replication
- [ ] Properties file
    - [ ] Validation
    - [ ] Using master properties file
- [ ] Partioner Hash function

## 🛺 Tech Stack
* Storage layer
    * [BoltDB](https://github.com/boltdb/bolt) and [BBoltDB](https://github.com/etcd-io/bbolt)
    * [BadgerDB](https://github.com/dgraph-io/badger)
    * [PebbleDB](https://github.com/cockroachdb/pebble)
* Consensus: [Hashicorp raft](https://github.com/hashicorp/raft)
* Consistent hashing: [Hashring](https://github.com/serialx/hashring)
* Storage format
    * [MessagePack](https://github.com/vmihailenco/msgpack)
    * [Avro](https://github.com/hamba/avro)
* Message passing: [GRPC](https://github.com/grpc/grpc-go)
* Clients
    * REST
    * CLI on rest
* and more ...

## ⚽ Contribute
Coming soon. Join our discord server till then