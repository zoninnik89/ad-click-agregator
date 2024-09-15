package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"
	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	"time"

	protoBuff "github.com/zoninnik89/commons/api"
)

type Service struct {
	store   store.StoreInterface
	gateway gateway.AdsGatewayInterface
}

func NewService(store store.StoreInterface, gateway gateway.AdsGatewayInterface) *Service {
	return &Service{store: store, gateway: gateway}
}

func (service *Service) GetClicksCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter := service.store.GetCount(request.AdId)

	//TO DO: add check if Ad exists

	return counter.ToProto(), nil
}

func (service *Service) StoreClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {
	var timestamp int64 = service.GenerateTS()
	service.store.AddClick(request.AdID)

	click := &protoBuff.Click{
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
func (service *Service) GenerateTS() int64 {
	return time.Now().Unix()
}
