#!/bin/bash

# Build proto files
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  models/jobmodels/job.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  components/network/network.proto

# Run the testcases
go test ./...

# Build the binary from source
go build