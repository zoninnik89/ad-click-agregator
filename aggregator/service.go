package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"

	protoBuff "github.com/zoninnik89/commons/api"
)

type Service struct {
	store   ClickStore
	gateway gateway.AdsGateway
}

func NewService(store ClickStore, gateway gateway.AdsGateway) *Service {
	return &Service{store: store, gateway: gateway}
}

func (service *Service) GetClicksCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter, err := service.store.Get(ctx, request.AdId)
	if err != nil {
		return nil, err
	}
	return counter.ToProto(), nil
}

func (service *Service) StoreClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {
	var timestamp int32 = 999999 // add gen func
	id, err := service.store.StoreClick(ctx, Click{
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

func (service *Service) ValidateClick(ctx context.Context, request *protoBuff.SendClickRequest) (bool, error) {
	validityCheck, err := service.gateway.CheckIfAdIsValid(ctx, request)
	if err != nil {
		return false, err
	}
	if validityCheck.Valid == true {
		return true, nil
	}
	return false, nil
}
