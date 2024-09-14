package gateway

import (
	"context"
	protoBuff "github.com/zoninnik89/commons/api"
	"github.com/zoninnik89/commons/discovery"
	"log"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *Gateway {

	return &Gateway{registry}
}

func (gateway *Gateway) CheckIfAdIsValid(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.AdValidity, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "ads", gateway.registry)
	if err != nil {
		log.Fatalf("Failed to dial to server: %v", err)
	}
	defer conn.Close()

	client := protoBuff.NewAdsServiceClient(conn)

	validityRequest := &protoBuff.CheckAdIsValidRequest{
		AdId:         request.AdID,
		ImpressionId: request.ImpressionID,
	}
	res, err := client.CheckAdIsValid(ctx, validityRequest)

	return res, err
}
