package gateway

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
)

type AdsGatewayInterface interface {
	CheckIfAdIsValid(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.AdValidity, error)
}
