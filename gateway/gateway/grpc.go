package gateway

import (
	"context"
	"log"

	protoBuff "github.com/zoninnik89/commons/api"
	"github.com/zoninnik89/commons/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (gateway *Gateway) CreateAd(ctx context.Context, proto *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "ads", gateway.registry)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}

	client := protoBuff.NewAdsServiceClient(conn)

	return client.CreateAd(ctx, &protoBuff.CreateAdRequest{
		ID:           proto.ID,
		AdvertiserID: proto.AdvertiserID,
		Title:        proto.Title,
		AdURL:        proto.AdURL,
	})
}

func (gateway *Gateway) GetAd(ctx context.Context, advertiserID, adID string) (*protoBuff.Ad, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "ads", gateway.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	client := protoBuff.NewAdsServiceClient(conn)

	return client.GetAD(ctx, &protoBuff.GetAdRequest{
		AdvertiserID: advertiserID,
		AdID:         adID,
	})
}
