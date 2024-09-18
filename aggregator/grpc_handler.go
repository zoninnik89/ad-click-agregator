package main

import (
	"context"
	"github.com/prometheus/common/log"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/types"
	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAggregatorServiceServer
	service types.AggregatorService
}

func NewGrpcHandler(grpcServer *grpc.Server, service types.AggregatorService) {
	handler := &GrpcHandler{
		service: service,
	}
	protoBuff.RegisterAggregatorServiceServer(grpcServer, handler)
}

func (handler *GrpcHandler) SendClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {

	click, err := handler.service.SendClick(ctx, request)
	if err != nil {
		log.Errorf("click storing failed with error: %v", err)
		return nil, err
	}

	log.Infof("click for adID: %v successfully stored successfully created at: %v", click.AdID, click.Timestamp)
	return click, nil
}

func (handler *GrpcHandler) GetClickCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter, err := handler.service.GetClickCounter(ctx, request)
	if err != nil {
		return nil, err
	}
	return counter, nil
}
