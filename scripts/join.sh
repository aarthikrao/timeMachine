# Set the number of nodes in the cluster and the leader node
num_nodes=$1
leader_ip="localhost"
leader_port=8001

# Join each node to the leader node
for i in $(seq 2 $num_nodes); do
  node_port=$((8100 + $i ))
  node_details='{"node_id":"node'$i'","raft_address": "localhost:'$node_port'"}'
  printf "node$i at localhost:$node_port : "
  curl -X POST -H "Content-Type: application/json" -d "$node_details" "http://localhost:8001/cluster/join"
  printf "\n"
done

printf "\nAll nodes joined to leader node at http://$leader_ip:$leader_port.\n"
