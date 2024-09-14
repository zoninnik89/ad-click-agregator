package main

import (
	"context"
	"github.com/google/uuid"
	protoBuff "github.com/zoninnik89/commons/api"
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	"github.com/zoninnik89/ad-click-aggregator/ads/gateway"
	_ "time"
)

type Service struct {
	store AdsStore
	gateway.AdGateway
	cache Cache
}

func NewService(store AdsStore, gateway gateway.AdGateway, cache Cache) *Service {
	return &Service{store, gateway, cache}
}

func (service *Service) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := service.store.Get(ctx, request.AdID, request.AdvertiserID)
	if err != nil {
		return nil, err
	}
	ad.ImpressionID, err = service.generateImpressionID(request.AdID, request.AdvertiserID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ad.ToProto(), nil
}

func (service *Service) CreateAd(ctx context.Context, request *protoBuff.CreateAdRequest) (*protoBuff.Ad, error) {
	id, err := service.store.Create(ctx, Ad{
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	})

	if err != nil {
		return nil, err
	}

	ad := &protoBuff.Ad{
		ID:           id.Hex(),
		AdvertiserID: request.AdvertiserID,
		Title:        request.Title,
		AdURL:        request.AdURL,
	}

	return ad, nil
}

type AdClaims struct {
	AdID         string `json:"adID"`
	AdvertiserID string `json:"advertiserID"`
	ImpressionID string `json:"impressionID"`
	jwt.RegisteredClaims
}

func (service *Service) generateImpressionID(adID, advertiserID string) (string, error) {
	impressionID := uuid.New().String()

	claims := AdClaims{
		AdID:         adID,
		AdvertiserID: advertiserID,
		ImpressionID: impressionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("secret")) // add vault later
	if err != nil {
		return "", err
	}
	service.cache.Put(signedToken)
	return signedToken, nil
}
