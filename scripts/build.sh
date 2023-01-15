#!/bin/bash

# Build proto files
protoc -I=models/jobmodels --go_out=models/jobmodels models/jobmodels/job_request.proto

# Run the testcases
go test ./...

# Build the binary from source
go build