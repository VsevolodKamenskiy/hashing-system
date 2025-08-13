package main

import (
	""
	"google.golang.org/grpc"
	"log"
	"net"
	"service1/internal/server"
	"service1/proto/hasherpb"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv := &server.Server{}

	hasherpb.RegisterHasherServiceServer(grpcServer, srv)

}
