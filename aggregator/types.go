package main

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
)

type AggregatorService interface {
	StoreClick(context.Context, *protoBuff.SendClickRequest) (*protoBuff.Click, error)
	GetClicksCounter(context.Context, *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error)
	ValidateClick(ctx context.Context, request *protoBuff.SendClickRequest) (bool, error)
}

type ClickCounter struct {
	adId        string
	totalClicks int32
}

type Click struct {
	ClickID    string
	AdID       string
	Timestamp  int32
	IsAccepted bool
}

func (clickCounter *ClickCounter) ToProto() *protoBuff.ClickCounter {
	return &protoBuff.ClickCounter{
		AdID:        clickCounter.adId,
		TotalClicks: clickCounter.totalClicks,
	}
}

func (click *Click) ToProto() *protoBuff.Click {
	return &protoBuff.Click{
		ClickID:   click.ClickID,
		AdID:      click.AdID,
		Timestamp: click.Timestamp,
	}
}
