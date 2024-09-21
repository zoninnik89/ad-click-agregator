package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/zoninnik89/ad-click-aggregator/ads/logging"
	protoBuff "github.com/zoninnik89/commons/api"
	"go.uber.org/zap"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	_ "time"
)

type Service struct {
	store  AdsStore
	logger *zap.SugaredLogger
}

func NewService(store AdsStore) *Service {
	l := logging.GetLogger().Sugar()
	return &Service{store: store, logger: l}
}

func (s *Service) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := s.store.Get(ctx, request.AdID, request.AdvertiserID)
	if err != nil {
		return nil, err
	}

	ad.ImpressionID, err = s.generateImpressionID(request.AdID)
	if err != nil {
		return nil, err
	}

	return ad.ToProto(), nil
}

func (s *Service) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	id, err := s.store.Create(ctx, Ad{
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	})

	if err != nil {
		return nil, err
	}

	ad := &protoBuff.Ad{
		AdID:         id.Hex(),
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	}

	return ad, nil
}

type AdClaims struct {
	AdID         string `json:"adID"`
	ImpressionID string `json:"impressionID"`
	jwt.RegisteredClaims
}

func (s *Service) generateImpressionID(adID string) (string, error) {
	impressionID := uuid.New().String()

	claims := AdClaims{
		AdID:         adID,
		ImpressionID: impressionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("secret")) // add vault later
	if err != nil {
		s.logger.Infow("Unable to sign impressionID", "adID", adID, "err", err)
		return "", err
	}

	s.logger.Infow("Generated impressionID", "impressionID", claims.ImpressionID, "at", claims.IssuedAt, "expiration time", claims.ExpiresAt)

	return signedToken, nil
}
