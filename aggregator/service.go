package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/common/log"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"
	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/types"
	protoBuff "github.com/zoninnik89/commons/api"
	"strings"
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

func (service *Service) ConsumeClick(ctx context.Context, consumer *kafka.Consumer) (*types.Click, error) {

	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		panic(err)
	}
	msgSlice := strings.Split(string(msg.Value), ",") // adID, impressionID, ts
	adID, impressionID, ts := msgSlice[0], msgSlice[1], msgSlice[2]

	_, err = service.CheckAdIsValid(ctx, adID, impressionID)
	if err != nil {
		return nil, err
	}

	service.store.AddClick(adID)
	service.cache.Put(impressionID)

	return &types.Click{AdID: adID, Timestamp: ts}, nil
}

//func generateTS() int64 {
//	currentTime := time.Now().Unix()
//	return currentTime
//}

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
