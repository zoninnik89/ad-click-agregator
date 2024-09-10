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

func (handler *GrpcHandler) CreateAd(ctx context.Context, protoBuff *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	log.Println("New ad created!")
	ad := handler.service.CreateAd(ctx, protoBuff)
	return ad, nil
}
