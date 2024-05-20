#!/bin/bash

# Ensure three arguments are provided: ip:port, shards, and replicas
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <ip:port> <shards> <replicas>"
  exit 1
fi

# Extract ip:port, shards, and replicas from command-line arguments
IP_PORT="$1"
SHARDS="$2"
REPLICAS="$3"

# Construct the target URL with the ip:port
URL="http://$IP_PORT/cluster/configure"

# Construct the JSON data in a multiline format for easy editing
DATA='{
  "shards":'"$SHARDS"',
  "replicas":'"$REPLICAS"'
}'

# Here you can easily add more parameters to the JSON by following the format
# For example, to add another_param:
# DATA=$(echo "$DATA" | jq '. + {another_param: "value"}')

# Send the HTTP POST request using cURL
curl --header "Content-Type: application/json" \
     --request POST \
     --data "$DATA" \
     "$URL"
