# Time Machine 🐓
A distributed, fault tolerant scheduler that can potentially scale to millions of jobs. 

The idea is to build it with a storage layer based on LSM-tree implementation, consistent hashing for load balancing, and raft for consensus.

## Stack
* [Badger](https://github.com/dgraph-io/badger)
* [Hashicorp raft](https://github.com/hashicorp/raft)
* [Hashring](https://github.com/serialx/hashring)
* and more ...
