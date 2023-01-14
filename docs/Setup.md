# 👩‍🏭 How to set up time machine

### Build the binary from code
```bash
❯ go test ./...
❯ go build
```

### Clean and data and raft folders
```bash
❯ # ./scripts/clean-create.sh num_nodes
❯ ./scripts/clean-create.sh 5
```

### Spawn multiple nodes
```bash
❯ # ./scripts/spawn.sh num_nodes is_bootstrap 
❯ ./scripts/spawn.sh 5 true 
```
Here the `spawn.sh` script accepts the number of nodes to spawn and true specifies if we should start in bootstrap mode. Bootstrap mode is necessary only when starting the cluster for the first time.

### Create a cluster
```bash
❯ # ./scripts/join.sh num_nodes
❯ ./scripts/join.sh 5
```
This script uses `curl` command to request the `node1` to accept `node2`, `node3`, `node4` and `node5` as followers.

### Check status
```bash
❯ # ./scripts/status.sh num_nodes
❯ ./scripts/status.sh 5

| No | Leader | Host:Port      | Health |
|----+--------+----------------+--------|
| 1  |   ✅   | 127.0.0.1:8001 |   🟢   |
| 2  |        | 127.0.0.1:8002 |   🟢   |
| 3  |        | 127.0.0.1:8003 |   🟢   |
| 4  |        | 127.0.0.1:8004 |   🟢   |
| 5  |        | 127.0.0.1:8005 |   🟢   |

Health and status checks completed for 5 nodes.
```
If you check the cluster status before [forming a cluster](#create-a-cluster), you will find that all the nodes we spawned are leaders in bootstrap mode

### Kill all nodes
```bash
pkill -f ./timeMachine
```