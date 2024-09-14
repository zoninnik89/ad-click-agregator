package main

import (
	"context"
	"fmt"
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
	cache Cache
}

func NewService(store AdsStore, cache Cache) *Service {
	return &Service{store, cache}
}

func (service *Service) GetAd(ctx context.Context, request *protoBuff.GetAdRequest) (*protoBuff.Ad, error) {
	ad, err := service.store.Get(ctx, request.AdID, request.AdvertiserID)
	if err != nil {
		return nil, err
	}
	ad.ImpressionID, err = service.generateImpressionID(request.AdID)
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
	service.cache.Put(signedToken)
	return signedToken, nil
}

func (service *Service) CheckAd(ctx context.Context, request *protoBuff.CheckAdIsValidRequest) (*protoBuff.AdValidity, error) {
	adValidity := &protoBuff.AdValidity{
		Valid: false,
	}

	if exists := service.cache.Get(request.ImpressionId); !exists {
		return adValidity, nil
	}

	parsedToken, err := jwt.Parse(request.ImpressionId, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return "secret", nil
	})

	// Handle parsing errors
	if err != nil {
		log.Println("Error parsing token:", err)
		return adValidity, fmt.Errorf("error parsing token: %v", err)
	}

	// Check if the token is valid
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Access specific fields in the JWT claims
		AdID := claims["AdID"]
		adValidity.Valid = AdID == request.AdId
		return adValidity, nil
	} else {
		return adValidity, fmt.Errorf("token is not valid")
	}
}
