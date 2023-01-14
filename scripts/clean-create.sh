#!/bin/bash
# This script was generated with the help of ChatGPT

# Set the number of nodes in the cluster
num_nodes=$1

# Remove the contents of data and raft directory for each node
rm -rf data/*

for i in $(seq 1 $num_nodes); do
  data_dir="data/node${i}/data"
  raft_dir="data/node${i}/raft"
  mkdir -p $data_dir
  mkdir -p $raft_dir
done

printf "Cleaned and created data and raft folders for $num_nodes nodes.\n"
