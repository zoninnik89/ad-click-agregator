package gateway

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/ads/logging"
	protoBuff "github.com/zoninnik89/commons/api"
	"github.com/zoninnik89/commons/discovery"
	"go.uber.org/zap"
)

type Gateway struct {
	registry discovery.Registry
	logger   *zap.SugaredLogger
}

func NewGRPCGateway(r discovery.Registry) *Gateway {
	return &Gateway{registry: r, logger: logging.GetLogger().Sugar()}
}

func (g *Gateway) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	g.logger.Infow("Starting connection with Ads service")
	conn, err := discovery.ServiceConnection(context.Background(), "ads", g.registry)
	if err != nil {
		g.logger.Fatalw("failed to dial server", "err", err)
		return nil, err
	}
	g.logger.Infof("Connection success")

	client := protoBuff.NewAdsServiceClient(conn)
	res, err := client.CreateAd(ctx, &protoBuff.CreateAdRequest{
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	})

	return res, nil
}

func (g *Gateway) GetAd(ctx context.Context, advertiserID, adID string) (*protoBuff.Ad, error) {
	g.logger.Infow("Starting connection with Ads service")
	conn, err := discovery.ServiceConnection(context.Background(), "ads", g.registry)

	if err != nil {
		g.logger.Fatalw("failed to dial server", "err", err)
		return nil, err
	}
	g.logger.Infof("Connection success")
	client := protoBuff.NewAdsServiceClient(conn)
	adRequest := &protoBuff.GetAdRequest{
		AdvertiserID: advertiserID,
		AdID:         adID,
	}
	result, err := client.GetAd(ctx, adRequest)

	return result, err
}

func (g *Gateway) GetClickCounter(ctx context.Context, adID string) (*protoBuff.ClickCounter, error) {
	g.logger.Infow("Starting connection with Aggregator service")
	conn, err := discovery.ServiceConnection(context.Background(), "aggregator", g.registry)
	if err != nil {
		g.logger.Fatalw("failed to dial server", "err", err)
	}

	client := protoBuff.NewAggregatorServiceClient(conn)
	return client.GetClickCounter(ctx, &protoBuff.GetClicksCounterRequest{
		AdId: adID,
	})
}
