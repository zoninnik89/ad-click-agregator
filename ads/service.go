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

func (service *Service) GetAd(ctx context.Context, protoBuff *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := service.store.Get(ctx, protoBuff.AdID, protoBuff.AdvertiserID)
	if err != nil {
		return nil, err
	}

	return ad.ToProto(), nil
}

func (service *Service) CreateAd(ctx context.Context, protoBuff *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	id, err := service.store.Create(ctx, Ad{
		AdvertiserID: protoBuff.AdvertiserID,
		Title:        protoBuff.Title,
		AdURL:        protoBuff.AdURL,
	})

	if err != nil {
		return nil, err
	}

	ad := &protoBuff.Ad{
		ID:           id.Hex(),
		AdvertiserID: protoBuff.AdvertiserID,
		Title:        protoBuff.Title,
		AdURL:        protoBuff.AdURL,
	}

	return ad, nil
}
