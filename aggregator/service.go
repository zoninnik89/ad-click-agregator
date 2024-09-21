package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/logging"
	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/types"
	protoBuff "github.com/zoninnik89/commons/api"
	"go.uber.org/zap"
	"strings"
)

type Service struct {
	store  store.StoreInterface
	cache  store.CacheInterface
	logger *zap.SugaredLogger
}

func NewService(store store.StoreInterface, cache store.CacheInterface) *Service {
	l := logging.GetLogger().Sugar()
	return &Service{store: store, cache: cache, logger: l}
}

func (s *Service) GetClickCounter(ctx context.Context, request *protoBuff.GetClicksCounterRequest) (*protoBuff.ClickCounter, error) {
	counter := s.store.GetCount(request.AdId)

	return counter.ToProto(), nil
}

func (s *Service) ConsumeClick(ctx context.Context, consumer *kafka.Consumer) (*types.Click, error) {
	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		s.logger.Fatalw("Failed to read message", "err", err)
	}
	msgSlice := strings.Split(string(msg.Value), ",") // adID, impressionID, ts
	adID, impressionID, ts := msgSlice[0], msgSlice[1], msgSlice[2]

	_, err = s.checkImpressionIDisValid(ctx, adID, impressionID)
	if err != nil {
		s.logger.Errorw("ImpressionID is not valid for the given adID", "adID", adID, "impressionID", impressionID, "err", err)
		return nil, err
	}

	s.store.AddClick(adID)
	s.cache.Put(impressionID)
	s.logger.Infow("Click was added to the store and to the cache", "adID", adID, "impressionID", impressionID, "at", ts)

	return &types.Click{AdID: adID, Timestamp: ts}, nil
}

func (s *Service) checkImpressionIDisValid(ctx context.Context, adID, tkn string) (bool, error) {
	s.logger.Infow("Checking impression ID", "adID", adID, "tkn", tkn)
	parsedToken, err := jwt.Parse(tkn, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})

	// Handle parsing errors
	if err != nil {
		return false, fmt.Errorf("error parsing token: %v", err)
	}

	// Check if the token is valid
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Access specific fields in the JWT claims
		adIDFromToken, adIDOk := claims["AdID"].(string) // Type assertion to string
		if !adIDOk {
			// If it's not directly a string, attempt to convert it
			adIDFromToken = fmt.Sprintf("%v", claims["adID"])
			s.logger.Infof("adID from the token was converted to: %v", adIDFromToken)
		}

		impressionIDFromToken, impressionIDOk := claims["impressionID"].(string) // Type assertion to string
		if !impressionIDOk {
			// If it's not directly a string, attempt to convert it
			impressionIDFromToken = fmt.Sprintf("%v", claims["ImpressionID"])
			s.logger.Infof("impressionID from the token was converted to: %v", impressionIDFromToken)
		}

		if impressionIDUsed := s.cache.Get(impressionIDFromToken); impressionIDUsed {
			return false, fmt.Errorf("impressionID: %v already exists", impressionIDFromToken)
		}
		if adID != adIDFromToken {
			return false, fmt.Errorf("adID from request: %v doesn't match adID from token: %v", adID, adIDFromToken)
		}
	}

	return true, nil
}

//func generateTS() int64 {
//	currentTime := time.Now().Unix()
//	return currentTime
//}
