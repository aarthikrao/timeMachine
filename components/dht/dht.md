# Distributed hash table implementation
We use custom DHT implementation to locate the data in the cluster of nodes. For more info, refer [dht.go interface](dht.go).

### Finding a key
to obtain the location of a key and its replica, 
* Assume a circle or a ring with `shards` number of nodes.
* Hash the key using `xxhash` hashing algorithm
* Distribute the slots across node in a mod based manner.

For example 
 - 
 Shard count: 12, replicas: 3

| Node  | Leader for Shards | Replica for Shards |
|-------|-------------------|--------------------|
| node0 | 0, 3, 6, 9        | 1, 4, 7, 10, 2, 5, 8, 11 |
| node1 | 1, 4, 7, 10       | 0, 3, 6, 9, 2, 5, 8, 11 |
| node2 | 2, 5, 8, 11       | 0, 3, 6, 9, 1, 4, 7, 10 |

### Why xxhash?

xxhash is an extremely fast non-cryptographic hash algorithm, working at speeds close to RAM limits. It's well-suited for hashing large amounts of data quickly, making it an ideal choice for applications requiring high-speed data processing and distribution.