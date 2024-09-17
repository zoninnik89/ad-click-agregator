package main

import (
	"context"
	"github.com/google/uuid"
	protoBuff "github.com/zoninnik89/commons/api"
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	_ "time"
)

type Service struct {
	store AdsStore
}

func NewService(store AdsStore) *Service {
	return &Service{store}
}

func (service *Service) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := service.store.Get(ctx, request.AdID, request.AdvertiserID)
	if err != nil {
		return nil, err
	}
	ad.ImpressionID, err = service.generateImpressionID(request.AdID)
	if err != nil {
		log.Printf("Error generating impression ID: %v\n", err)
		return nil, err
	}
	log.Printf("Impression id was successfully generated for ad: %v with value: %v\n", ad.AdID, ad.ImpressionID)

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

func (service *Service) generateImpressionID(adID string) (string, error) {
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
		return "", err
	}

	log.Printf("Generated impression ID: %v at: %v with expiration time: %v", claims.ImpressionID, claims.IssuedAt, claims.ExpiresAt)

	return signedToken, nil
}
