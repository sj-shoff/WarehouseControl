package app

import "google.golang.org/grpc"

type grpcAdapter struct {
	server *grpc.Server
}
