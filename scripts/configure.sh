#!/bin/bash

# Set the target URL
URL="http://localhost:8001/cluster/configure"

# Check if slot_per_node_count is provided as a command-line argument
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <slot_per_node_count>"
  exit 1
fi

# Extract slot_per_node_count from command-line arguments
SLOT_PER_NODE_COUNT="$1"

# Construct the JSON data
DATA="{\"slot_per_node_count\":$SLOT_PER_NODE_COUNT}"

# Set additional configurable parameters if needed
# EXAMPLE: DATA+='{"another_param":"value"}'

# Send the HTTP POST request using cURL
curl --header "Content-Type: application/json" \
     --request POST \
     --data "$DATA" \
     "$URL"
