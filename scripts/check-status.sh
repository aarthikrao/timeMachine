#!/bin/bash
# This script was generated with the help of ChatGPT

# Set the number of nodes in the cluster
num_nodes=$1

# Call the health and status API of each node and store the output in an array
printf "| %s | %s | %s      | %s \n" "No" "Leader" "Host:Port" "Health |" 
echo "|----+--------+----------------+--------|"
for i in $(seq 1 $num_nodes); do
  node_ip=127.0.0.1
  node_port=$((8000 + $i ))

  # Check the health
  health_output=$(curl -s "http://$node_ip:$node_port/health" | jq -r '.status')
  if [[ -n $health_output ]] && [ $health_output == "ok" ]; then
      health_output="🟢"
  else
      health_output="🔴"
  fi

  # Check the status
  status_output=$(curl -s "http://$node_ip:$node_port/cluster" | jq -r '.state')
    if [[ -n $status_output ]] && [ $status_output == "Leader" ]; then
      status_output="✅"
  else
      status_output="  "
  fi

  # Print the output
  printf "| %d  |   %s   | %s:%d |   %s   |\n" "$i" "$status_output" "$node_ip" "$node_port" "$health_output" 
done | column -t -s "\t"

printf "\nHealth and status checks completed for $num_nodes nodes.\n"
