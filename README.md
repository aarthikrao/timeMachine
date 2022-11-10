# Time Machine DB 🐓
[![Discord](https://img.shields.io/badge/Discord-%235865F2.svg?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/pDGNPj3dTM) 
![Status](https://img.shields.io/badge/Status-Ideation-ffb3ff?style=for-the-badge)

A distributed, fault tolerant scheduler database that can potentially scale to millions of jobs. 

The idea is to build it with a storage layer based on B+tree or LSM-tree implementation, consistent hashing for load balancing, and raft for consensus.

## 🧬 Documentation
- [Purpose](./docs/Purpose.md)
    - [Why are we building this ?](./docs/Purpose.md#-why-are-we-building-this)
    - [What does it take ?](./docs/Purpose.md#🚜-what-does-it-take)
- [Architecture](./docs/Architecture.md)
- [Developer APIs](./docs/DevAPI.md)
    - [Job APIs](./docs/DevAPI.md#-job-apis)
    - [Route APIs](./docs/DevAPI.md#-route-apis)
- [TODO](./docs/TODO.md)

## Roadmap 
To be revealed soon 🍄

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


## 🔭 Also read
* [Leto](https://github.com/yongman/leto) - A KV DB built on hasicorp raft and badger db
* [RocksDB](https://github.com/facebook/rocksdb) - A KV DB by facebook