package main

import (
	"context"

	"github.com/zoninnik89/ad-click-aggregator/ads/gateway"
	protoBuff "github.com/zoninnik89/commons/api"
)

type Service struct {
	store AdsStore
	gateway.AdGateway
}

func NewService(store AdsStore, gateway gateway.AdGateway) *Service {
	return &Service{store, gateway}
}

func (service *Service) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := service.store.Get(ctx, request.AdID, request.AdvertiserID)
	if err != nil {
		return nil, err
	}

	return ad.ToProto(), nil
}

func (service *Service) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	id, err := service.store.Create(ctx, Ad{
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	})

	if err != nil {
		return nil, err
	}

	ad := &protoBuff.Ad{
		ID:           id.Hex(),
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	}

	return ad, nil
}
