package main

import (
	"context"
	"log"

	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAdsServiceServer

	service AdsService
}

func NewGrpcHandler(grpcServer *grpc.Server, service AdsService) {
	handler := &GrpcHandler{
		service: service,
	}
	protoBuff.RegisterAdsServiceServer(grpcServer, handler)
}

func (handler *GrpcHandler) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	log.Printf("Got to get ad")
	return handler.service.GetAd(ctx, request)
}

func (handler *GrpcHandler) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	log.Println("New ad created!")
	ad, err := handler.service.CreateAd(ctx, request)
	if err != nil {
		return nil, err
	}
	return ad, nil
}
