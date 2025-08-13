package server

import (
	"context"
	"service1/pkg/hasher"
	"service1/proto/hasherpb"
)

type Server struct {
	hasherpb.UnimplementedHasherServiceServer
}

func (s *Server) CalculateHashes(ctx context.Context, req *hasherpb.HashRequest) (*hasherpb.HashResponse, error) {
	hashes := hasher.HashStringsParallel(req.GetStrings())
	return &hasherpb.HashResponse{Hashes: hashes}, nil
}
