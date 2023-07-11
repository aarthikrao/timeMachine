# Components in timeMachine node

![components](../docs/images/components.png)

## Components
### [Client process](./client/client.go)
The client process is responsible for providing abstractions related to clients. The REST and GRPC servers use this to expose data to other nodes and clients.

### [Concensus](./concensus/Concensus.md)
The consensus module is implemented using the RAFT protocol. It is used for handling the configuration requirements of the time machine cluster. We can store and share information such as the DHT structure, route information and other cluster-related information. The RAFT algorithm elects a leader(or a leaseholder) for consensus who will also act as the leader of the time machine cluster.

### [Distributed hash table](./dht/dht.md)
The distributed hash table is responsible for maintaining the location of all the data in a time machine cluster. This is initialised once the raft cluster is formed. You can read more about this in the DHT documentation.

### [Data store manager](./datastore/datastore.go)
The datastore manager handles the abstraction layer for the datastore in the node. It implements the same JobStore interface.

### [Data stores](./datastore/)
These are the storage engines(implemented using BBlot) that hold all the data of a particular vnode. During the migration of the vnode, the entire datastore is copied to the new location.

### [Connection manager](../process/connectionmanager/connection_manager.go)
The connection manager handles all connections for the time machine node. It uses GRPC for communication with other nodes.

### [Node manager](../process/nodemanager/node_manager.go)
The node manager is the central location for all the processing in the time machine node. It also handles the initialisation of the DHT, node and cluster.
