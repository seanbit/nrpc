package main

import (
	"context"
	"github.com/seanbit/nrpc"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/protos/test"
)

type Service struct {
	component.Base
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) RequestHandle(ctx context.Context, args *test.SomeRequest) (*test.SomeResponse, error) {
	logger := nrpc.GetDefaultLoggerFromCtx(ctx) // The default logger contains a requestId, the route being executed and the sessionId
	logger.WithFields(map[string]interface{}{"args": args}).Info("RequestHandle called")
	// Do nothing. This is just an example of how pipelines can be helpful
	return &test.SomeResponse{
		Code: "ok",
		Resp: args.String(),
	}, nil
}
