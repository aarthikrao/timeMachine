package network

import "google.golang.org/grpc"

type networkHandler struct {
	client JobStoreClient
}

func CreateConnection(addr string) (*networkHandler, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock())

	if err != nil {
		return nil, err
	}

	return &networkHandler{
		client: NewJobStoreClient(conn),
	}, nil
}

