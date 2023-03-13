# Distributed hash table implementation
We use custom DHT implementation to locate the data in the cluster of nodes. For more info, refer [dht.go interface](dht.go).

### Finding a key
to obtain the location of a key and its replica, 
* Assume a circle with `slotCount` number of nodes.
* Hash the key using `xxhash` hashing algorithm
* Distribute the slots across node in a contiguous manner. 
```
[0:node1 1:node1 2:node1 3:node1 4:node2 5:node2 6:node2 7:node2 8:node3 9:node3 10:node3 11:node3]
SlotCount = 12, NodeCount = 3
```
* The replica for this key will be in the diagonally opposite slot of the circle.
    Mathematically, we add the first replica slot value plus half the `slotCount` to get the diagonally opposite slot number.
* We get the physical nodeIDs from the `slotVsNode` map

### Proposing a slot to migrate
This algoritm is yet to be finalised.