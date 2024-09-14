package gateway

import (
	"context"

	protoBuff "github.com/zoninnik89/commons/api"
)

type AdGateway interface {
	CreateAd(context.Context, *protoBuff.CreateAdRequest) (*protoBuff.Ad, error)
	GetAd(ctx context.Context, advertiserID, adID string) (*protoBuff.Ad, error)
	SendClick(context.Context, *protoBuff.SendClickRequest) (*protoBuff.Click, error)
	GetClickCounter(ctx context.Context, adID string) (*protoBuff.ClickCounter, error)
}
