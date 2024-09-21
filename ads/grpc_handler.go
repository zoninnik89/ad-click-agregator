package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/ads/logging"
	protoBuff "github.com/zoninnik89/commons/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	protoBuff.UnimplementedAdsServiceServer
	logger  *zap.SugaredLogger
	service AdsService
}

func NewGrpcHandler(grpcServer *grpc.Server, service AdsService) {
	handler := &GrpcHandler{
		logger:  logging.GetLogger().Sugar(),
		service: service,
	}
	protoBuff.RegisterAdsServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	h.logger.Infow("Received GetAd request", "adID", request.AdID)
	r, err := h.service.GetAd(ctx, request)
	if err != nil {
		h.logger.Errorw("Failed to get ad", "adID", request.AdID, "error", err)
	}
	h.logger.Infow("Ad was successfully retrieved from DB ", "adID", request.AdID, "response", r)
	return r, nil
}

func (h *GrpcHandler) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	h.logger.Infow("Received CreatAD request", "adID", request.AdURL, "advertiserID", request.AdvertiserID, "title", request.Title)
	ad, err := h.service.CreateAd(ctx, request)
	if err != nil {
		h.logger.Errorw("Error creating the ad", "adID", request.AdURL)
		return nil, err
	}
	h.logger.Infow("Successfully created the ad", "adID", request.AdURL, "response", ad)
	return ad, nil
}
