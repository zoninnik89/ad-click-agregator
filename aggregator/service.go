package main

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/common/log"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"
	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	"time"

	protoBuff "github.com/zoninnik89/commons/api"
)

type Service struct {
	store   store.StoreInterface
	gateway gateway.AdsGatewayInterface
	cache   store.CacheInterface
}

func NewService(store store.StoreInterface, gateway gateway.AdsGatewayInterface, cache store.CacheInterface) *Service {
	return &Service{store: store, gateway: gateway, cache: cache}
}

func (service *Service) GetClickCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter := service.store.GetCount(request.AdId)

	//TO DO: add check if Ad exists

	return counter.ToProto(), nil
}

func (service *Service) SendClick(ctx context.Context, request *protoBuff.SendClickRequest) (*protoBuff.Click, error) {
	var timestamp int64 = generateTS()

	validityCheck, err := service.CheckAdIsValid(ctx, request.AdID, request.ImpressionID)
	if err != nil {
		log.Errorf("error during click validation: %v", err)
		return nil, err
	}

	service.store.AddClick(request.AdID)
	service.cache.Put(request.ImpressionID)

	log.Infof("click for ad: %v with timestamp: %v was added to the store", request.AdID, timestamp)

	click := &protoBuff.Click{
		AdID:       request.AdID,
		Timestamp:  timestamp,
		IsAccepted: validityCheck,
	}

	return click, nil
}

func generateTS() int64 {
	currentTime := time.Now().Unix()
	return currentTime
}

func (service *Service) CheckAdIsValid(ctx context.Context, adID, tkn string) (bool, error) {

	parsedToken, err := jwt.Parse(tkn, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})

	// Handle parsing errors
	if err != nil {
		log.Errorf("error while parsing token: %v", err)
		return false, fmt.Errorf("error parsing token: %v", err)
	}

	// Check if the token is valid
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Access specific fields in the JWT claims
		adIDFromToken, adIDOk := claims["AdID"].(string) // Type assertion to string
		if !adIDOk {
			// If it's not directly a string, attempt to convert it
			adIDFromToken = fmt.Sprintf("%v", claims["adID"])
			log.Infof("adID was converted to: %v", adIDFromToken)
		}

		impressionIDFromToken, impressionIDOk := claims["impressionID"].(string) // Type assertion to string
		if !impressionIDOk {
			// If it's not directly a string, attempt to convert it
			impressionIDFromToken = fmt.Sprintf("%v", claims["ImpressionID"])
			log.Infof("impressionID was converted to: %v", impressionIDFromToken)
		}

		if impressionIDUsed := service.cache.Get(impressionIDFromToken); impressionIDUsed {
			return false, fmt.Errorf("ImpressionID: %v already exists", impressionIDFromToken)
		}
		if adID != adIDFromToken {
			return false, fmt.Errorf("adID from request: %v doesn't match adID from token: %v", adID, adIDFromToken)
		}
	}

	return true, nil
}
