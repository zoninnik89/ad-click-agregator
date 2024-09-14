package main

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAggregatorServiceServer
	service AggregatorService
}

func NewGrpcHandler(grpcServer *grpc.Server, service AggregatorService) {
	handler := &GrpcHandler{
		service: service,
	}
	protoBuff.RegisterAggregatorServiceServer(grpcServer, handler)
}

func (handler *GrpcHandler) SendClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {
	clickIsValid, err := handler.service.ValidateClick(ctx, request)
	if !clickIsValid {
		return nil, err
	}
	click, err := handler.service.StoreClick(ctx, request)
	if err != nil {

		return nil, err
	}
	return click, nil
}

func (handler *GrpcHandler) GetClicksCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter, err := handler.service.GetClicksCounter(ctx, request)
	if err != nil {
		return nil, err
	}
	return counter, nil
}
