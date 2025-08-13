package server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"service1/pkg/hasher"
	"service1/proto/hasherpb"
)

type Server struct {
	hasherpb.UnimplementedHasherServiceServer
	Log         *logrus.Logger
	ShutdownCtx context.Context
}

func (s *Server) CalculateHashes(reqCtx context.Context, req *hasherpb.HashRequest) (*hasherpb.HashResponse, error) {
	ctx, cancel := context.WithCancel(reqCtx)
	defer cancel()

	// прокидываем глобальную отмену (shutdown) в этот child
	if s.ShutdownCtx != nil {
		go func() {
			<-s.ShutdownCtx.Done()
			cancel()
		}()
	}

	log := GetLoggerFromCtx(ctx, s.Log)
	strs := req.GetStrings()

	log.WithField("count", len(strs)).Info("hash fan-in start")

	hashes, err := hasher.HashStringsParallel(ctx, strs)

	if err != nil {
		werr := errors.WithStack(err)
		log.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).Error("hash fan-in failed")
		return nil, err
	}

	log.WithField("count", len(hashes)).Info("hash fan-in done")
	return &hasherpb.HashResponse{Hashes: hashes}, nil
}
