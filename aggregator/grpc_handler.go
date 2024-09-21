package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/logging"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/types"
	protoBuff "github.com/zoninnik89/commons/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAggregatorServiceServer
	service types.AggregatorService
	logger  *zap.SugaredLogger
}

func NewGrpcHandler(grpcServer *grpc.Server, service types.AggregatorService) {
	handler := &GrpcHandler{
		service: service,
		logger:  logging.GetLogger().Sugar(),
	}
	protoBuff.RegisterAggregatorServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) GetClickCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	h.logger.Infow("GetClickCounter request received", "adID", request.AdId)
	counter, err := h.service.GetClickCounter(ctx, request)
	if err != nil {
		return nil, err
	}
	return counter, nil
}
