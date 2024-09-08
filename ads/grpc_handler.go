package main

import (
	"context"
	"log"

	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAdsServiceServer
}

func NewGrpcHandler(grpcServer *grpc.Server) {
	handler := &GrpcHandler{}
	protoBuff.RegisterAdsServiceServer(grpcServer, handler)
}

func (handler *GrpcHandler) CreateAd(context.Context, *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	log.Println("New ad created!")
	ad := &protoBuff.Ad{
		ID: "42",
	}
	return ad, nil
}
