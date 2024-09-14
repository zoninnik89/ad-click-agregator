package gateway

import (
	"context"
	//protoBuff "github.com/zoninnik89/commons/api"
	"github.com/zoninnik89/commons/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *Gateway {

	return &Gateway{registry}
}

func (gateway Gateway) CheckIfAdIsValid(ctx context.Context, adID string) (bool, error) {

}
