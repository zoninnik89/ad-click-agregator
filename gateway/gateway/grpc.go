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

func (gateway *Gateway) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "ads", gateway.registry)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}

	client := protoBuff.NewAdsServiceClient(conn)

	return client.CreateAd(ctx, &protoBuff.CreateAdRequest{
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	})
}

func (gateway *Gateway) GetAd(ctx context.Context, advertiserID, adID string) (*protoBuff.Ad, error) {
	log.Printf("starting the connection")
	conn, err := discovery.ServiceConnection(context.Background(), "ads", gateway.registry)

	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	log.Printf("connection succeded")
	client := protoBuff.NewAdsServiceClient(conn)
	adRequest := &protoBuff.GetAdRequest{
		AdvertiserID: advertiserID,
		AdID:         adID,
	}
	result, err := client.GetAd(ctx, adRequest)

	return result, err
}

func (gateway *Gateway) GetClickCounter(ctx context.Context, adID string) (*protoBuff.ClickCounter, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "aggregator", gateway.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	client := protoBuff.NewAggregatorServiceClient(conn)
	return client.GetClickCounter(ctx, &protoBuff.GetClicksCounterRequest{
		AdId: adID,
	})
}
