#!/bin/bash
# This script was generated using ChatGPT

# Set the number of nodes in the cluster
num_nodes=$1

# Call the health API of each node and store the output in an array
printf "\nno\thost:port\toutput\n"
for i in $(seq 1 $num_nodes); do
  node_ip=127.0.0.1
  node_port=$((8000 + $i ))
  health_output=$(curl -s "http://$node_ip:$node_port/health")
  printf "$i\t$node_ip:$node_port\t${health_output}\n"
done | column -t 

printf "\nHealth checks completed for $num_nodes nodes.\n"
