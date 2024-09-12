package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"

	_ "github.com/zoninnik89/ad-click-aggregator/ads/gateway"
	protoBuff "github.com/zoninnik89/commons/api"
)

type Service struct {
	store ClickStore
	gateway.AggregatorGateway
}

func NewService(store ClickStore, gateway gateway.AggregatorGateway) *Service {
	return &Service{store, gateway}
}

func (service *Service) GetClickCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter, err := service.store.Get(ctx, request.AdId)
	if err != nil {
		return nil, err
	}
	return counter.ToProto(), nil
}

func (service *Service) SendClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {
	var timestamp int32 = 999999 // add gen func
	id, err := service.store.AddClick(ctx, Click{
		AdID: request.AdID,
	})
	if err != nil {
		return nil, err
	}

	click := &protoBuff.Click{
		ClickID:    id,
		AdID:       request.AdID,
		Timestamp:  timestamp,
		IsAccepted: true,
	}
	return click, nil
}
