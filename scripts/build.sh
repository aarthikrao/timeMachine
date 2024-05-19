#!/bin/bash

# Build proto files
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  models/jobmodels/job.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  components/network/network.proto

# Get all go modules
go mod tidy

# Run the testcases
go test ./...

# Build the binary from source
go build -o timeMachine ./cmd/server
go build -o timeMachineCli ./cmd/cli
go build -o tester ./cmd/integrationtest
