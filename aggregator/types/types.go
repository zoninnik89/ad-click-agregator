package types

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
)

type AggregatorService interface {
	SendClick(context.Context, *protoBuff.SendClickRequest) (*protoBuff.Click, error)
	GetClickCounter(context.Context, *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error)
}

type ClickCounter struct {
	AdId        string
	TotalClicks int32
}

type Click struct {
	ClickID    string
	AdID       string
	Timestamp  int64
	IsAccepted bool
}

func (clickCounter *ClickCounter) ToProto() *protoBuff.ClickCounter {
	return &protoBuff.ClickCounter{
		AdID:        clickCounter.AdId,
		TotalClicks: clickCounter.TotalClicks,
	}
}

func (click *Click) ToProto() *protoBuff.Click {
	return &protoBuff.Click{
		AdID:       click.AdID,
		Timestamp:  click.Timestamp,
		IsAccepted: click.IsAccepted,
	}
}
