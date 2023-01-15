#!/bin/bash
# This script was generated with the help of ChatGPT

# Set the number of nodes in the cluster
num_nodes=$1

# Set to true if you want to start each node of the cluster in a bootstrap mode.
# This is only done initally and timeMachine will throw an error if the cluster already existed
bootstrap=$2

# Start the first node
node1_ip=127.0.0.1
node1_port=8000
echo "Starting node 1 at $node1_ip:$node1_port"
./timeMachine --serverID=node1 --raftPort=8101 --httpPort=8001 --bootstrap=$bootstrap &

# Start the remaining nodes
for i in $(seq 2 $num_nodes); do
  echo "Starting node$i with Raft at $node_ip:810$i and HTTP at 800$i"
  ./timeMachine --serverID=node$i --raftPort=810$i --httpPort=800$i --bootstrap=$bootstrap &
done

echo "Local cluster started with $num_nodes nodes."
