package types

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
)

type AggregatorService interface {
	GetClickCounter(context.Context, *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error)
}

type ClickCounter struct {
	AdId        string
	TotalClicks int32
}

type Click struct {
	AdID      string
	Timestamp string
}

func (clickCounter *ClickCounter) ToProto() *protoBuff.ClickCounter {
	return &protoBuff.ClickCounter{
		AdID:        clickCounter.AdId,
		TotalClicks: clickCounter.TotalClicks,
	}
}
