package gateway

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
)

type AdsGateway interface {
	CheckIfAdIsValid(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.AdValidity, error)
}
